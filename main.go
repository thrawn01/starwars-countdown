package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"starwars-countdown/middleware"
)

var opts struct {
	Bind     string `short:"b" long:"bind" env:"SWCD_BIND_ADDR" default:"localhost:8080"`
	ImageDir string `short:"i" long:"image-dir" env:"SWCD_IMAGE_DIR" default:"images/"`
	Debug    bool   `short:"d" long:"debug"`
}

type CountDown struct {
	ImageDir string
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
	countDown := &CountDown{opts.ImageDir}

	// Add our routes
	router := httprouter.New()
	//router.GET("/", countDown.Index)
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
	path := fmt.Sprintf("public%s", params.ByName("path"))

	// Determine our content type by file extension
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		log.Info("Unable to determine mime type for ", path)
		resp.Header().Set("Content-Type", "text/html")
	} else {
		resp.Header().Set("Content-Type", ctype)
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
		http.Error(resp, "500 Internal Server Error", http.StatusForbidden)
		return
	}
	defer fd.Close()

	// Write the entire file back to the client
	io.Copy(resp, fd)
}
