package main

// tpl loads templates from a filesystem
// If the template was loaded in development mode and the template has
// changed since last loaded, it is reloaded

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"time"
)

// tpl records data about a template.
// Note that the relPath needs to be relative to the root of fileSystem.
type tpl struct {
	inDevelopment bool
	fileSystem    fs.FS
	path          string
	updated       time.Time
	tpl           *template.Template
}

// TplParse a template file, optionally recording its last updated time.
// The relPath needs to be relative to the root of fileSystem.
func TplParse(dev bool, fileSystem fs.FS, relPath string) (tpl, error) {
	t := tpl{inDevelopment: dev, fileSystem: fileSystem, path: relPath}
	err := t.parse()
	return t, err
}

// parse parses the template the first time, and then again on Execute
// if the tpl is inDevelopment and the file has been modified
func (t *tpl) parse() error {
	if !t.inDevelopment && !t.updated.IsZero() {
		return nil // return early if not indevelopment, and initialised
	}
	if t.inDevelopment {
		stat, err := fs.Stat(t.fileSystem, t.path)
		if err != nil {
			return fmt.Errorf("could not get template details %s: %v", t.path, err)
		}
		updated := stat.ModTime()
		if !updated.After(t.updated) {
			return nil
		}
		t.updated = updated
	}
	var err error
	t.tpl, err = template.ParseFS(t.fileSystem, t.path)
	if err != nil {
		return fmt.Errorf("could not load template %s: %v", t.path, err)
	}
	return nil
}

// Execute checks the file by reparsing and then executes the template
func (t tpl) Execute(w io.Writer, data any) error {
	err := t.parse()
	if err != nil {
		return err
	}
	return t.tpl.Execute(w, data)
}
