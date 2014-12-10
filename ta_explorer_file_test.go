package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileTaExplorer(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	reg := NewFileTaExplorer(path, 0)
	if _, err := reg.(*taExplorer).base.Put("list", testServExpTree); err != nil {
		t.Fatal(err)
	}

	testTaExplorer(t, reg)
}

func TestFileTaExplorerStamp(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	reg := NewFileTaExplorer(path, 0)
	if _, err := reg.(*taExplorer).base.Put("list", testServExpTree); err != nil {
		t.Fatal(err)
	}

	testTaExplorerStamp(t, reg)
}
