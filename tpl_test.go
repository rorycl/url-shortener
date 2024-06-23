package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writer(file *os.File, contents string) error {
	_, err := fmt.Fprint(file, contents)
	if err != nil {
		return err
	}
	return file.Sync()
}

// Test reloading of development templates
func TestTplReloading(t *testing.T) {

	for _, inDevelopment := range []bool{true, false} {

		dir := os.DirFS("/tmp")
		file, err := os.CreateTemp("/tmp", "tpl_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(file.Name())

		// write first version
		err = writer(file, "---\ntitle {{ .Title }}\n")
		if err != nil {
			t.Fatal(err)
		}

		// resync
		tpl, err := TplParse(inDevelopment, dir, filepath.Base(file.Name()))
		update1 := tpl.updated

		// write second version
		time.Sleep(time.Millisecond * 5)
		err = writer(file, "Update\n")
		if err != nil {
			t.Fatal(err)
		}
		err = tpl.parse()
		if err != nil {
			t.Fatal(err)
		}
		update2 := tpl.updated
		if inDevelopment && !update2.After(update1) {
			t.Errorf("inDevelopment %v not after %v", update1, update2)
		} else if !inDevelopment && !update2.Equal(update1) {
			t.Errorf("production %v should equal %v", update2, update1)
		}

		err = tpl.Execute(os.Stdout, map[string]string{"Title": fmt.Sprintf("%t", inDevelopment)})
		if err != nil {
			t.Fatal(err)
		}
		update3 := tpl.updated
		if !update3.Equal(update2) {
			t.Errorf("%v not after %v", update1, update2)
		}
	}
}
