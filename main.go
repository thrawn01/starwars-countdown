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
	"starwars-countdown/middleware"
	log "starwars-countdown/vendor/_nuts/github.com/Sirupsen/logrus"
	flags "starwars-countdown/vendor/_nuts/github.com/jessevdk/go-flags"
	"starwars-countdown/vendor/_nuts/github.com/julienschmidt/httprouter"
	"starwars-countdown/vendor/_nuts/github.com/justinas/alice"
	"strings"
)

var opts struct {
	BindAddress string `short:"b" long:"bind-address" env:"SWCD_BIND_ADDR" default:"localhost:8080"`
	ImageDir    string `short:"i" long:"image-dir" env:"SWCD_IMAGE_DIR" description:"Location of the images within the public-dir"  default:"images/"`
	PublicDir   string `short:"p" long:"public-dir" env:"SWCD_PUBLIC_DIR" description:"The directory where index.html lives" default:"public/"`
	OutputIndex bool   `short:"o" long:"output-index" description:"Print index.html to stdout and exit"`
	Debug       bool   `short:"d" long:"debug"`
}

type TemplateContent struct {
	Images []string
}

type CountDown struct {
	CachedIndexPage bytes.Buffer
	PublicDir       string
}

func RenderIndexPage(publicDir string, imageDir string) bytes.Buffer {

	// Collect a list of images to slide show
	pathToImageDir := filepath.Join(publicDir, imageDir)
	files, err := ioutil.ReadDir(pathToImageDir)
	if err != nil {
		log.Fatal(err)
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
				log.Fatal(err)
			}
			// Add the url to our image list
			url := strings.Trim(string(uri), "\n")
			log.Debug("Found lnk: ", url)
			imageContents = append(imageContents, string(url))
		} else {
			// Add a relative path to the image list
			relativePathToImage := filepath.Join(imageDir, file.Name())
			log.Debug("Found image: ", relativePathToImage)
			imageContents = append(imageContents, relativePathToImage)
		}
	}

	// Read in index.html for caching
	indexFile, err := ioutil.ReadFile(filepath.Join(publicDir, "index.html"))
	if err != nil {
		log.Fatal(err)
	}

	// Render index.html
	renderer := template.New("Index template")
	renderer, err = renderer.Parse(string(indexFile))
	if err != nil {
		log.Error(err)
	}
	templateContent := &TemplateContent{imageContents}
	var parsedIndexPage bytes.Buffer
	renderer.Execute(&parsedIndexPage, templateContent)
	return parsedIndexPage
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(-1)
	}

	if opts.Debug {
		log.Printf("Debug enabled")
		log.SetLevel(log.DebugLevel)
	}

	// Render the index.html page
	indexPage := RenderIndexPage(opts.PublicDir, opts.ImageDir)
	// Do we output and exit?
	if opts.OutputIndex {
		io.Copy(os.Stdout, &indexPage)
		return
	}

	// Create a new instance
	countDown := &CountDown{indexPage, opts.PublicDir}
	// And init the server
	server := countDown.NewServer()
	// Add our middleware chain
	chain := alice.New(middleware.NewRequestLogger(500), middleware.Timeout).Then(server)
	// Listen and serve requests
	log.Info("Listening for requests on ", opts.BindAddress)
	err = http.ListenAndServe(opts.BindAddress, chain)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (self *CountDown) NewServer() http.Handler {
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

func (self *CountDown) ServeFiles(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	// Redirect requests for '/' to '/index.html'
	if req.URL.Path == "/" {
		log.Debug("Redirect to '/index.html'")
		self.Redirect(resp, req, "/index.html")
		return
	}

	path := filepath.Join(self.PublicDir, params.ByName("path"))

	// Determine our content type by file extension
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		log.Debug("Unable to determine mime type for ", path)
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
