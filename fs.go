package main

import (
	"fmt"
	"io/fs"
	"os"
)

// fs implements a simple filesystem abstraction for accessing files for
// static web serving and templates suitable for embedding or running
// live off a local machine during development.

// NewFileSystem returns a new fileSystem
func NewFileSystem(inDevelopment bool, path string, ebed fs.FS) (fs.FS, error) {

	var f fs.FS

	// use embedded filesystem if not in development
	if !inDevelopment {
		var err error
		f, err = fs.Sub(ebed, path)
		return f, err
	}

	// otherwise use direct path (allows live reloading of templates, etc)
	if !dirOK(path) {
		return f, fmt.Errorf("path %s could not be mounted", path)
	}
	f = os.DirFS(path)
	return f, nil
}

// dirOK checks if a directory path is ok
func dirOK(d string) bool {
	if d == "" {
		return false
	}
	if _, err := os.Stat(d); os.IsNotExist(err) {
		return false
	}
	return true
}
