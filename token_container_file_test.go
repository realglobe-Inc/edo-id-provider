package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"os"
	"testing"
)

var testPriKey crypto.PrivateKey

func init() {
	var err error
	testPriKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
}

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

	testTokenContainer(t, newFileTokenContainer(10, "https://example.com", testPriKey, "", "RS256", testSavDur, path, expiPath, testStaleDur, testCaExpiDur))
}
