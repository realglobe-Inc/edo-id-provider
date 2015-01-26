package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"testing"
	"time"
)

var testTaPriKey crypto.PrivateKey
var testTaPubKey crypto.PublicKey

var testTa *ta

func init() {
	priKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	testTaPriKey = priKey
	testTaPubKey = &priKey.PublicKey

	testTa = newTa(
		"testta",
		"testtaname",
		map[string]bool{
			"https://testta.example.org/":             true,
			"https://testta.example.org/redirect/uri": true,
		},
		map[string]crypto.PublicKey{
			"": testTaPubKey,
		})
	testTa.Upd = testTa.Upd.Add(-(time.Duration(testTa.Upd.Nanosecond()) % time.Millisecond)) // mongodb の粒度がミリ秒のため。
}

func testTaContainer(t *testing.T, taCont taContainer) {
	if ta_, err := taCont.get(testTa.id()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ta_, testTa) {
		t.Error(ta_, testTa)
	}

	if ta_, err := taCont.get(testTa.id() + "a"); err != nil {
		t.Fatal(err)
	} else if ta_ != nil {
		t.Error(ta_)
	}
}
