package main

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/braintree/manners"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/thrawn01/args"
	"github.com/thrawn01/starwars-countdown/middleware"
)

type TemplateContent struct {
	Images []string
}

type CountDown struct {
	CachedIndexPage bytes.Buffer
	PublicDir       string
}

func RenderIndexPage(publicDir string, imageDir string) bytes.Buffer {

	// Collect a list of images for slide show
	pathToImageDir := filepath.Join(publicDir, imageDir)
	files, err := ioutil.ReadDir(pathToImageDir)
	if err != nil {
		logrus.Fatal(err)
	}

	// Decide what images are links and what images are local
	imageContents := make([]string, 0)
	for _, file := range files {
		// If the file name ends in .lnk
		if strings.HasSuffix(file.Name(), ".lnk") {
			// Read the file and retrieve the link to the image
			pathToLinkFile := filepath.Join(pathToImageDir, file.Name())
			uri, err := ioutil.ReadFile(pathToLinkFile)
			if err != nil {
				logrus.Fatal(err)
			}
			// Add the url to our image list
			url := strings.Trim(string(uri), "\n")
			logrus.Debug("Found lnk: ", url)
			imageContents = append(imageContents, string(url))
		} else {
			// Add a relative path to the image list
			relativePathToImage := filepath.Join(imageDir, file.Name())
			logrus.Debug("Found image: ", relativePathToImage)
			imageContents = append(imageContents, relativePathToImage)
		}
	}

	// Read in index.html for caching
	indexHtml, err := ioutil.ReadFile(filepath.Join(publicDir, "index.html"))
	if err != nil {
		logrus.Fatal(err)
	}

	// Render index.html
	renderer := template.New("Index template")
	renderer, err = renderer.Parse(string(indexHtml))
	if err != nil {
		logrus.Error(err)
	}
	templateContent := &TemplateContent{imageContents}
	var parsedIndexPage bytes.Buffer
	renderer.Execute(&parsedIndexPage, templateContent)
	return parsedIndexPage
}

func main() {
	parser := args.NewParser(args.EnvPrefix("SWCD_"))
	parser.AddOption("--bind").Env("BIND_ADDR").Default("localhost:8080").
		Help("interface to listen on")
	parser.AddOption("--image-dir").Env("IMAGE_DIR").Default("images/").
		Help("Location of the images within the public-dir")
	parser.AddOption("--public-dir").Env("PUBLIC_DIR").Default("public/").
		Help("The directory where index.html lives")
	parser.AddOption("--output-index").Alias("-o").IsTrue().
		Help("Print index.html to stdout and exit")
	parser.AddOption("--debug").Alias("-o").IsTrue().
		Help("Print index.html to stdout and exit")

	opts := parser.ParseArgsSimple(nil)

	if opts.Bool("debug") {
		logrus.Printf("Debug enabled")
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Render the index.html page
	indexPage := RenderIndexPage(opts.String("public-dir"), opts.String("image-dir"))
	// Do we output and exit?
	if opts.Bool("output-index") {
		io.Copy(os.Stdout, &indexPage)
		return
	}

	// Create a new instance
	countDown := &CountDown{indexPage, opts.String("public-dir")}
	// And init the server
	router := countDown.NewRouter()
	// Add our middleware chain
	chain := alice.New(middleware.NewRequestLogger(500),
		middleware.NewThrottle(),
		middleware.Timeout).Then(router)

	server := manners.NewWithServer(&http.Server{
		Addr:    opts.String("bind"),
		Handler: chain,
	})

	// Catch SIGINT Gracefully so we don't drop any active http requests
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, os.Kill)
		sig := <-signalChan
		logrus.Infof("Captured %v. Exiting...", sig)
		server.Close()
	}()

	// Listen and serve requests
	logrus.Info("Listening for requests on ", opts.String("bind"))
	err := server.ListenAndServe()
	if err != nil {
		logrus.Fatal("ListenAndServe: ", err)
	}
}

func (self *CountDown) NewRouter() http.Handler {
	// Add our routes
	router := httprouter.New()
	router.GET("/*path", self.ServeFiles)
	return router
}

func (self *CountDown) Redirect(resp http.ResponseWriter, req *http.Request, newPath string) {
	if query := req.URL.RawQuery; query != "" {
		newPath += "?" + query
	}
	resp.Header().Set("Location", newPath)
	resp.WriteHeader(http.StatusMovedPermanently)
}

// TODO: Replace some of this code with
// http.Handle("/", http.FileServer(http.Dir("public")))
func (self *CountDown) ServeFiles(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	// Redirect requests for '/' to '/index.html'
	if req.URL.Path == "/" {
		logrus.Debug("Redirect to '/index.html'")
		self.Redirect(resp, req, "/index.html")
		return
	}

	path := filepath.Join(self.PublicDir, params.ByName("path"))

	// Determine our content type by file extension
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		logrus.Debug("Unable to determine mime type for ", path)
		resp.Header().Set("Content-Type", "text/html")
	} else {
		resp.Header().Set("Content-Type", ctype)
	}

	// If this is the index return the cached index page
	if req.URL.Path == "/index.html" {
		readable := bytes.NewReader(self.CachedIndexPage.Bytes())
		io.Copy(resp, readable)
		return
	}

	// Open the requested file
	fd, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(resp, "404 page not found", http.StatusNotFound)
			return
		}
		if os.IsPermission(err) {
			http.Error(resp, "403 forbidden", http.StatusForbidden)
			return
		}
		http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	// Write the entire file back to the client
	io.Copy(resp, fd)
}
