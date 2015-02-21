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

	// 動作に影響は出ないはず。
	testSavDur    = 15 * time.Millisecond
	testStaleDur  = 10 * time.Millisecond
	testCaExpiDur = 5 * time.Millisecond

	// 動作に影響あり。
	// Go の GC が 10ms くらいは時間を使うと言っているので、それ以上に。
	testCodExpiDur   = 20 * time.Millisecond
	testTokExpiDur   = 20 * time.Millisecond
	testIdTokExpiDur = 20 * time.Millisecond
	testSessExpiDur  = 20 * time.Millisecond
	testTicDur       = 20 * time.Millisecond

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
		map[string]interface{}{
			testTaKid: testTaPubKey,
		})
	testTa.Upd = testTa.Upd.Add(-(time.Duration(testTa.Upd.Nanosecond()) % time.Millisecond)) // mongodb の粒度がミリ秒のため。
}
