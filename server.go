package main

import (
	"fmt"
	"html"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
)

// go:embed templates
var templates fs.FS
var templatePath = "templates"

// go:embed static
var static fs.FS
var staticPath = "static"

// go:embed data
var data fs.FS
var dataPath = "data"
var dataFile = "pd-short-urls.csv"

const defaultPort = "8000"
const defaultAddr = "0.0.0.0"

func (s *server) serve() {

	log.Printf("serving on %s", s.FullAddress())

	r := http.NewServeMux()

	// mount static filepath
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/",
			http.FileServer(http.FS(s.static)),
		),
	)

	// routes
	r.HandleFunc("GET /{$}", home)
	r.HandleFunc("GET /{shortURL}", s.redirector)
	r.HandleFunc("GET /{anyURL...}", notFound)

	// middleware
	logging := func(handler http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(os.Stdout, handler)
	}
	recovery := func(handler http.Handler) http.Handler {
		return handlers.RecoveryHandler()(handler)
	}
	chainedHandlers := alice.New(recovery, logging, r)

	// configure server options
	httpServer := &http.Server{
		Addr:    s.FullAddress(),
		Handler: chainedHandlers,
		// timeouts and limits
		MaxHeaderBytes:    1 << 17, // ~125k
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	log.Printf("serving on %s", addr, port)

	err := listenAndServe(httpServer)
	if err != nil {
		log.Printf("fatal server error: %v", err)
	}
}

func (s server) FullAddress() string {
	return strings.Join([]string{s.addr, s.port}, ":")
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "home")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	anyURL := r.PathValue("anyURL")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "The url %s is invalid", html.EscapeString(anyURL))
}

func (s *server) redirector(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	longURL, ok := s.urlMap[shortURL]
	if ok {
		fmt.Fprintf(w, "%s -> \n%s", shortURL, longURL)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "The url %s was not found", html.EscapeString(shortURL))
}

func main() {
	s, err := newServer(true, "0.0.0.0", "8001")
	if err != nil {
		log.Fatal(err)
	}
	s.serve()
}

// server describes the main settings for the server
type server struct {
	urlMap        map[string]string // the map of short to full urls
	inDevelopment bool              // use the file system or embedded resources
	addr          string
	port          string
	templates     fs.FS // fs.FS to templates
	static        fs.FS // fs.FS to static resources
	data          fs.FS // fs.FS to csv file of short to full urls
}

// newServer creates a new server
func newServer(dev bool, addr, port string) (*server, error) {
	var err error
	if addr == "" {
		addr = defaultAddr
	}
	if port == "" {
		port = defaultPort
	}
	s := server{
		inDevelopment: dev,
		addr:          addr,
		port:          port,
	}

	// attach file systems
	s.templates, err = NewFileSystem(s.inDevelopment, templatePath, templates)
	if err != nil {
		return &s, fmt.Errorf("could not attach template filesystem: %v", err)
	}
	s.static, err = NewFileSystem(s.inDevelopment, staticPath, static)
	if err != nil {
		return &s, fmt.Errorf("could not attach template filesystem: %v", err)
	}
	s.data, err = NewFileSystem(s.inDevelopment, dataPath, data)
	if err != nil {
		return &s, fmt.Errorf("could not attach template filesystem: %v", err)
	}

	// load urls
	dataFile, err := s.data.Open(dataFile)
	if err != nil {
		return &s, fmt.Errorf("could not open data file: %v", err)
	}
	s.urlMap, err = urls(dataFile)
	if err != nil {
		return &s, fmt.Errorf("could not load urls: %v", err)
	}

	return &s, nil
}
