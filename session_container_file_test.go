package main

import (
	"io/ioutil"
	"os"
	"testing"
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

	testSessionContainer(t, newFileSessionContainer(10, path, expiPath, testStaleDur, testCaExpiDur))
}
