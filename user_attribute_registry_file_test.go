package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileUserAttributeRegistry(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	reg := NewFileUserAttributeRegistry(path, 0)
	if _, err := reg.(*userAttributeRegistry).base.Put(testUsrUuid+"/"+testAttrName, testAttr); err != nil {
		t.Fatal(err)
	}

	testUserAttributeRegistry(t, reg)
}
