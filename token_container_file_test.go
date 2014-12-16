package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestFileTokenContainer(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	expiPath, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(expiPath)

	testTokenContainer(t, newFileTokenContainer(10, time.Second, time.Second, path, expiPath, 0, 0))
}
