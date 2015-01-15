package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestFileSessionContainer(t *testing.T) {
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

	testSessionContainer(t, newFileSessionContainer(10, 20*time.Millisecond, path, expiPath, 0, 0))
}
