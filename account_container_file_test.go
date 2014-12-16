package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileAccountContainer(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	namePath, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(namePath)

	if buff, err := json.Marshal(testAcc); err != nil {
		t.Fatal(err)
	} else if err := ioutil.WriteFile(filepath.Join(path, keyToEscapedJsonPath(testAcc.Id)), buff, filePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(path, keyToEscapedJsonPath(testAcc.Id)), filepath.Join(namePath, keyToEscapedJsonPath(testAcc.Name))); err != nil {
		t.Fatal(err)
	}

	testAccountContainer(t, newFileAccountContainer(path, namePath, 0, 0))
}
