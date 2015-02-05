package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"time"
)

const (
	filePerm = 0644
)

const (
	testLabel = "edo-test"

	testIdLen = 5
	testUiUri = "/html"

	testSavDur    = 15 * time.Millisecond
	testStaleDur  = 5 * time.Millisecond
	testCaExpiDur = time.Millisecond

	testCodExpiDur   = 10 * time.Millisecond
	testTokExpiDur   = 10 * time.Millisecond
	testIdTokExpiDur = 10 * time.Millisecond
	testSessExpiDur  = 10 * time.Millisecond

	testTicDur = 20 * time.Millisecond

	testSigAlg = "RS256"
)

var testIdpPriKey crypto.PrivateKey
var testIdpPubKey crypto.PublicKey

func init() {
	priKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	testIdpPriKey = priKey
	testIdpPubKey = &priKey.PublicKey
}

const testTaKid = "testkey"

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
			testTaKid: testTaPubKey,
		})
	testTa.Upd = testTa.Upd.Add(-(time.Duration(testTa.Upd.Nanosecond()) % time.Millisecond)) // mongodb の粒度がミリ秒のため。
}
