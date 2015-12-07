package main

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"starwars-countdown/middleware"
	"strings"
)

var opts struct {
	Bind      string `short:"b" long:"bind" env:"SWCD_BIND_ADDR" default:"localhost:8080"`
	ImageDir  string `short:"i" long:"image-dir" env:"SWCD_IMAGE_DIR" default:"public/images/"`
	PublicDir string `short:"p" long:"public-dir" env:"SWCD_PUBLIC_DIR" default:"public"`
	Debug     bool   `short:"d" long:"debug"`
}

type ImageContent struct {
	Uri     string
	IsLocal bool
}

type TemplateContent struct {
	Images []ImageContent
}

type CountDown struct {
	CachedIndexPage bytes.Buffer
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Panic(err)
	}

	if opts.Debug {
		log.Printf("Debug Enabled")
		log.SetLevel(log.DebugLevel)
	}

	// Collect a list of images to slide show
	files, err := ioutil.ReadDir(opts.ImageDir)
	if err != nil {
		log.Fatal(err)
	}

	// Decide what images are links and what images are local
	imageContents := make([]ImageContent, 0)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".lnk") {
			uri, err := ioutil.ReadFile(filepath.Join(opts.ImageDir, file.Name()))
			if err != nil {
				log.Fatal(err)
			}
			urlEndpoint := strings.Trim(string(uri), "\n")
			log.Debug("lnk file: ", file.Name())
			imageContents = append(imageContents, ImageContent{string(urlEndpoint), false})
		} else {
			log.Debug("image file: ", file.Name())
			// Assume the file is an image
			imageContents = append(imageContents, ImageContent{file.Name(), true})
		}
	}

	// Pre-render and cache the index.html file
	indexFile, err := ioutil.ReadFile(filepath.Join(opts.PublicDir, "index.html"))
	if err != nil {
		log.Fatal(err)
	}

	renderer := template.New("Index template")
	renderer, err = renderer.Parse(string(indexFile))
	if err != nil {
		log.Error(err)
	}
	templateContent := &TemplateContent{imageContents}
	var parsedIndexPage bytes.Buffer
	renderer.Execute(&parsedIndexPage, templateContent)

	countDown := &CountDown{parsedIndexPage}

	// Add our routes
	router := httprouter.New()
	router.GET("/*path", countDown.ServeFiles)

	// Add our middleware chain
	chain := alice.New(middleware.NewRequestLogger(1000), middleware.Timeout).Then(router)

	// Listen and serve requests
	log.Printf("Listening for requests on %s", opts.Bind)
	http.ListenAndServe(opts.Bind, chain)
}

func Redirect(resp http.ResponseWriter, req *http.Request, newPath string) {
	if query := req.URL.RawQuery; query != "" {
		newPath += "?" + query
	}
	resp.Header().Set("Location", newPath)
	resp.WriteHeader(http.StatusMovedPermanently)
}

func (countDown *CountDown) ServeFiles(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	// Redirect requests for '/' to '/index.html'
	if req.URL.Path == "/" {
		log.Info("Redirect to '/index.html'")
		Redirect(resp, req, "/index.html")
		return
	}

	// TODO: Path to files should be configurable
	path := filepath.Join("public", params.ByName("path"))

	// Determine our content type by file extension
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		log.Info("Unable to determine mime type for ", path)
		resp.Header().Set("Content-Type", "text/html")
	} else {
		resp.Header().Set("Content-Type", ctype)
	}

	// If this is the index file parse the file
	if req.URL.Path == "/index.html" {
		io.Copy(resp, &countDown.CachedIndexPage)
		countDown.CachedIndexPage.Reset()
	}

	// Open the requested file
	fd, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(resp, "404 page not found", http.StatusNotFound)
		}
		if os.IsPermission(err) {
			http.Error(resp, "403 forbidden", http.StatusForbidden)
		}
		http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer fd.Close()

	// Write the entire file back to the client
	io.Copy(resp, fd)
}
