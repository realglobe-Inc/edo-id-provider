package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileTaContainer(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	if buff, err := marshalTa(testTa); err != nil {
		t.Fatal(err)
	} else if err := ioutil.WriteFile(filepath.Join(path, keyToEscapedJsonPath(testTa.id())), buff, filePerm); err != nil {
		t.Fatal(err)
	}

	testTaContainer(t, newFileTaContainer(path, testStaleDur, testCaExpiDur))
}
