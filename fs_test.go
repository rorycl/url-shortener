package main

import (
	"embed"
	"io/fs"
	"testing"
)

//go:embed data
var testData embed.FS

// TestFSMount tests to see if the static and templates filesystems can
// be mounted and read
func TestFSMount(t *testing.T) {

	testCases := []struct {
		name          string
		inDevelopment bool
		path          string
		ebd           embed.FS
	}{
		{"production", false, "data", testData}, // production embed fs doesn't need arguments
		{"development", true, "data", testData},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mount, err := NewFileSystem(tc.inDevelopment, tc.path, tc.ebd)
			if err != nil {
				t.Fatal(err)
			}

			d, err := fs.ReadDir(mount, ".")
			t.Log(d, err)

			_, err = fs.ReadFile(mount, "pd-short-urls.csv")
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDirOk(t *testing.T) {
	if dirOK("") {
		t.Errorf("empty dir should error")
	}
	if dirOK("nonexisting") {
		t.Errorf("nonexisting dir should error")
	}
	if !dirOK("data") {
		t.Errorf("data dir should not error")
	}
}
