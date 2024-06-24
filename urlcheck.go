package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// urlcheck checks if remote urls return 200 status

const httpWorkers = 12 // also goroutines to spin up
const httpTimeout time.Duration = 200 * time.Millisecond

var NonHTMLPageType error = errors.New("non html page type")

// getClient encapsulates an http.Client and the functions used against
// that client, which are parameterised to allow for convenient swapping
// out during testing
type getClient struct {
	client  *http.Client
	workers int
}

// NewGetClient initialises a new getClient.
func NewGetClient(workers int, timeout time.Duration) *getClient {
	g := getClient{}
	if workers == 0 {
		g.workers = httpWorkers
	} else {
		g.workers = workers
	}
	if timeout == 0 {
		timeout = httpTimeout
	}
	g.client = &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: g.workers,
		},
		Timeout: timeout,
	}
	return &g
}

// Check checks a set of urls, returning the count of processed errors
// until an error or completion
func (g *getClient) Check(urls []string) (count int, errorCount int) {

	type result struct {
		url    string
		status int
		err    error
	}

	getURL := func(urlChan <-chan string, results chan<- result) {
		for u := range urlChan {
			status, err := g.get(u)
			results <- result{u, status, err}
		}
	}

	urlChan := make(chan string, len(urls))
	resultChan := make(chan result)

	for range g.workers {
		go getURL(urlChan, resultChan)
	}

	for _, uu := range urls {
		urlChan <- uu
	}
	close(urlChan)

	for rr := range resultChan {
		count++
		if rr.err != nil {
			errorCount++
			fmt.Printf("%s\n   %v\n", rr.url, rr.err)
		}
		if rr.err == nil && rr.status != 200 {
			errorCount++
			fmt.Printf("%s\n   status %d\n", rr.url, rr.status)
		}
		if count == len(urls) {
			break
		}
	}
	return count, errorCount
}

// get gets a URL, reporting the status and erroring if the page is not
// of an html type
func (g *getClient) get(url string) (status int, err error) {
	resp, err := g.client.Get(url)
	if err != nil {
		return status, err
	}
	status = resp.StatusCode
	/* not needed if some urls are, for example, PDFs
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		return status, NonHTMLPageType
	}
	*/
	return status, nil
}
