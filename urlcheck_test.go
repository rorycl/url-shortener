package main

import (
	"fmt"
	"testing"
	"time"
)

func TestGet(t *testing.T) {

	g := NewGetClient(1, 200*time.Millisecond)

	tests := []struct {
		url    string
		status int
		isErr  bool
	}{
		{"https://www.google.com", 200, false},
		{"www.google.com", 0, true},
		// {"https://jsonplaceholder.typicode.com/todos/1", 200, true}, // non-html error
		{"https://www.theguardian.com/qwertyuiop", 404, false}, // 404

	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			status, err := g.get(tt.url)
			if err != nil && !tt.isErr {
				t.Errorf("unexpected error %v for %s", err, tt.url)
			}
			if err == nil {
				if got, want := status, tt.status; got != want {
					t.Errorf("unexpected status %d (want %d) for %s", status, tt.status, tt.url)
				}

			}
		})
	}
}

func TestGetURLs(t *testing.T) {

	tests := []struct {
		workers  int
		timeout  time.Duration
		urls     []string
		count    int
		errCount int
	}{
		{
			workers: 3,
			timeout: 350 * time.Millisecond,
			urls: []string{
				"https://www.google.com",
				"https://www.theguardian.com",
			},
			count:    2,
			errCount: 0,
		},
		{
			workers: 4,
			timeout: 350 * time.Millisecond,
			urls: []string{
				"https://www.google.com",
				"https://github.com",
				"https://www.330661ae-31b2-11ef-b7f6-2702c69d005d.ltd", // doesn't exist
			},
			count:    3,
			errCount: 1,
		},
		{
			workers: 1,
			timeout: 350 * time.Millisecond,
			urls: []string{
				"https://github.com",
				"https://www.google.com",
				"https://www.gov.uk/",
				"https://www.gov.uk//qwertyuiop", // 404
			},
			count:    3,
			errCount: 1,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			g := NewGetClient(tt.workers, tt.timeout)
			count, errCount := g.Check(tt.urls)
			if errCount != tt.errCount {
				t.Logf("error count got %d expected %d", errCount, tt.errCount)
			}
			if count != tt.count {
				t.Logf("count got %d expected %d", errCount, tt.errCount)
			}
		})
	}
}
