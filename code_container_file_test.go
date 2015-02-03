package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileCodeContainer(t *testing.T) {
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

	testCodeContainer(t, newFileCodeContainer(10, "", testSavDur, testTicDur, path, expiPath, testStaleDur, testCaExpiDur))
}
