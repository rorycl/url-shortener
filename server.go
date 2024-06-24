package main

import (
	"embed"
	_ "embed"
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

//go:embed templates
var templatesFS embed.FS
var templatePath = "templates"

//go:embed static
var staticFS embed.FS
var staticPath = "static"

//go:embed data
var dataFS embed.FS
var dataPath = "data"
var dataFile = "short-urls.csv"

// defaults
const defaultPort = "8000"
const defaultAddr = "0.0.0.0"

// run the server
func (s *server) serve() error {
	r := http.NewServeMux()

	// routes using go's new 1.22 routes
	r.HandleFunc("GET /{$}", s.home)
	r.HandleFunc("GET /{shortURL}", s.redirector)
	r.HandleFunc("GET /{anyURL...}", s.invalid)
	r.Handle("GET /static/", s.staticFiles())

	// middleware; consider throttling middleware too
	// gorilla mux middleware "Add" is nice also
	logging := func(handler http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(os.Stdout, handler)
	}
	recovery := func(handler http.Handler) http.Handler {
		return handlers.RecoveryHandler()(handler)
	}
	chainedHandlers := alice.New(recovery, logging).Then(r)

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

	log.Printf("Running server on %s", s.FullAddress())
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Printf("fatal server error: %v", err)
	}
	return err
}

// FullAddress makes a full address of the addr and port
func (s *server) FullAddress() string {
	return strings.Join([]string{s.addr, s.port}, ":")
}

// vals returns the long urls
func (s *server) vals() []string {
	vSlice := []string{}
	for _, v := range s.urlMap {
		vSlice = append(vSlice, v)
	}
	return vSlice
}

// staticFiles mounts the static file fs.FS
func (s *server) staticFiles() http.Handler {
	return http.StripPrefix("/"+staticPath+"/",
		http.FileServer(http.FS(s.static)),
	)
}

// home is a home page handler
func (s *server) home(w http.ResponseWriter, r *http.Request) {
	err := s.homeTpl.Execute(w, struct{ Title string }{"Home"})
	if err != nil {
		errorOutput(w, "home", err)
	}
}

// invalid is a 404 handler for invalid paths
func (s *server) invalid(w http.ResponseWriter, r *http.Request) {
	anyURL := r.PathValue("anyURL")
	vars := struct {
		Title, URL  string
		InvalidPath bool
	}{"Invalid Path", html.EscapeString(anyURL), true}
	w.WriteHeader(http.StatusNotFound)
	err := s.notFoundTpl.Execute(w, vars)
	if err != nil {
		errorOutput(w, "not found", err)
	}
}

// redirector is the main handler, which falls through to a 404 if no
// short url key can be found in s.urlMap. Otherwise the user is
// redirected with a 301 (StatusMovedPermanently) redirect.
func (s *server) redirector(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	longURL, ok := s.urlMap[shortURL]
	if ok {
		http.Redirect(w, r, longURL, http.StatusMovedPermanently)
		return
	}
	// short code not found
	vars := struct {
		Title, URL  string
		InvalidPath bool
	}{"Redirection not found", html.EscapeString(shortURL), false}
	w.WriteHeader(http.StatusNotFound)
	err := s.notFoundTpl.Execute(w, vars)
	if err != nil {
		errorOutput(w, "redirection not found", err)
	}
}

// server holds the main settings for the server
type server struct {
	urlMap        map[string]string // the map of short to full urls
	inDevelopment bool              // use the file system or embedded resources
	addr          string
	port          string
	templates     fs.FS // templates
	static        fs.FS // static resources
	data          fs.FS // csv file with short,full urls
	homeTpl       tpl
	notFoundTpl   tpl
	httpTimeout   time.Duration // http client timeout
	httpWorkers   int
}

// newServer creates a new server and attaches various resources
func newServer(dev bool, addr, port string, timeout time.Duration, workers int) (*server, error) {
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
		httpTimeout:   timeout,
		httpWorkers:   workers,
	}

	// attach file systems
	s.templates, err = NewFileSystem(s.inDevelopment, templatePath, templatesFS)
	if err != nil {
		return &s, fmt.Errorf("could not attach template filesystem: %v", err)
	}
	s.static, err = NewFileSystem(s.inDevelopment, staticPath, staticFS)
	if err != nil {
		return &s, fmt.Errorf("could not attach static filesystem: %v", err)
	}
	s.data, err = NewFileSystem(s.inDevelopment, dataPath, dataFS)
	if err != nil {
		return &s, fmt.Errorf("could not attach data filesystem: %v", err)
	}

	// templates
	s.homeTpl, err = TplParse(s.inDevelopment, s.templates, "home.html")
	if err != nil {
		return &s, fmt.Errorf("could not load home template: %v", err)
	}
	s.notFoundTpl, err = TplParse(s.inDevelopment, s.templates, "404.html")
	if err != nil {
		return &s, fmt.Errorf("could not load 404 template: %v", err)
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

	// verify urls if in development
	if s.inDevelopment {
		g := NewGetClient(s.httpWorkers, s.httpTimeout)
		count, errCount := g.Check(s.vals())
		fmt.Printf("url check reported %d errors in %d url checks\n", errCount, count)
	}

	return &s, nil
}

// errorOutput is a convenience func for reporting errors
func errorOutput(w http.ResponseWriter, source string, err error) {
	log.Printf("%s template error %v", source, err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "template writing problem at %s: %s", source, err.Error())
}
