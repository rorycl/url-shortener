package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestServerDevelopment(t *testing.T) {

	ns, err := newServer(true, "127.0.0.1", "8765", 200*time.Millisecond, 2)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err = ns.serve()
		if err != nil {
			// cannot use testing.T here
			fmt.Print(err) // maybe the port is in use?
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
		return
	}()

	var status int
	var body string
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name         string
		method       string
		url          string
		status       int
		bodyContains string
	}{
		{"home_okay", "GET", "http://127.0.0.1:8765/", 200, "URL Shortener"},
		{"home_post", "POST", "http://127.0.0.1:8765/", 405, ""},
		{"not found", "GET", "http://127.0.0.1:8765/abc", 404, "was not found"},
		{"invalid", "GET", "http://127.0.0.1:8765/abc/d", 404, "invalid path"},
		{"file ok", "GET", "http://127.0.0.1:8765/static/styles.css", 200, "margin"},
		{"file notok", "GET", "http://127.0.0.1:8765/static/nonsense", 404, "not"},
		{"redirect", "GET", "http://127.0.0.1:8765/dbd", 301, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body, err = httpClient(tt.method, tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := status, tt.status; got != want {
				t.Errorf("expected status %d got %d", want, got)
			}
			if (status == 200 || status == 404) && !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body does not contain %s", tt.bodyContains)
			}

		})
	}
}

// https://www.digitalocean.com/community/tutorials/how-to-make-http-requests-in-go
func httpClient(method, url string) (int, string, error) {

	var body string
	var status int
	var err error
	var request *http.Request

	// https://dev.to/fuchiao/http-redirect-for-golang-http-client-2i35
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 100 * time.Millisecond,
	}

	switch method {
	case "GET":
		request, err = http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return status, body, fmt.Errorf("get error: %v", err)
		}
	case "POST":
		request, err = http.NewRequest(http.MethodPost, url, strings.NewReader(""))
		if err != nil {
			return status, body, fmt.Errorf("get error: %v", err)
		}
	default:
		return status, body, fmt.Errorf("method %s not supported", method)
	}

	res, err := client.Do(request)
	if err != nil {
		return status, body, fmt.Errorf("Http %s request error at %s: %v\n", method, url, err)
	}
	defer res.Body.Close()
	status = res.StatusCode

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return status, body, fmt.Errorf("could not read body %v", err)
	}
	body = string(bodyBytes)

	return status, body, nil
}

// thisRWriter implements http.ResponseWriter
// https://stackoverflow.com/a/76861153
type thisRWriter struct {
	b      bytes.Buffer
	status int
}

func (t *thisRWriter) Header() http.Header { return http.Header{} }
func (t *thisRWriter) Write(data []byte) (int, error) {
	return t.b.Write(data)
}
func (t *thisRWriter) WriteHeader(status int) {
	t.status = status
}

func TestErrorOutput(t *testing.T) {
	trw := &thisRWriter{}
	errorOutput(trw, "tpl", errors.New("tpl1"))
	if got, want := string(trw.b.Bytes()), "template writing problem at tpl: tpl1"; got != want {
		t.Errorf("got %s != want %s", got, want)
	}
}
