package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileUserNameIndex(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	reg := NewFileUserNameIndex(path, 0)
	if _, err := reg.(*userNameIndex).base.Put(testUsrName, testUsrUuid); err != nil {
		t.Fatal(err)
	}
	testUserNameIndex(t, reg)
}
