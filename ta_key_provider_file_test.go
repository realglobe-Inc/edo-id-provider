package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileTaKeyProvider(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	reg := NewFileTaKeyProvider(path, 0, 0)
	if _, err := reg.(*taKeyProvider).base.Put(testServUuid, testPublicKey); err != nil {
		t.Fatal(err)
	}

	testTaKeyProvider(t, reg)
}

func TestFileTaKeyProviderStamp(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	reg := NewFileTaKeyProvider(path, 0, 0)
	if _, err := reg.(*taKeyProvider).base.Put(testServUuid, testPublicKey); err != nil {
		t.Fatal(err)
	}

	testTaKeyProvider(t, reg)
}
