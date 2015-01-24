package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileConsentContainer(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	testConsentContainer(t, newFileConsentContainer(path, testStaleDur, testCaExpiDur))
}
