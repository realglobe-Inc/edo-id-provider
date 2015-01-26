package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func init() {
	util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
}

const (
	testIdLen = 5
	testUiUri = "/html"

	testCodExpiDur   = 10 * time.Millisecond
	testTokExpiDur   = 10 * time.Millisecond
	testIdTokExpiDur = 10 * time.Millisecond
	testSessExpiDur  = 10 * time.Millisecond

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

func newTestSystem(selfId string) *system {
	uiPath, err := ioutil.TempDir("", testLabel)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(uiPath, selHtml), []byte{}, filePerm); err != nil {
		os.RemoveAll(uiPath)
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(uiPath, loginHtml), []byte{}, filePerm); err != nil {
		os.RemoveAll(uiPath)
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(uiPath, consHtml), []byte{}, filePerm); err != nil {
		os.RemoveAll(uiPath)
		panic(err)
	}
	return &system{
		selfId,
		false,
		testIdLen,
		testIdLen,
		testUiUri,
		uiPath,
		newMemoryTaContainer(testStaleDur, testCaExpiDur),
		newMemoryAccountContainer(testStaleDur, testCaExpiDur),
		newMemoryConsentContainer(testStaleDur, testCaExpiDur),
		newMemorySessionContainer(testIdLen, testStaleDur, testCaExpiDur),
		newMemoryCodeContainer(testIdLen, testSavDur, testStaleDur, testCaExpiDur),
		newMemoryTokenContainer(testIdLen, testSavDur, testStaleDur, testCaExpiDur),
		testCodExpiDur + 2*time.Second, // 以下、プロトコルを通すと粒度が秒になるため。
		testTokExpiDur + 2*time.Second,
		testIdTokExpiDur + 2*time.Second,
		testSessExpiDur + 2*time.Second,
		testSigAlg,
		"",
		testIdpPriKey,
	}
}

// 起動しただけでパニックを起こさないこと。
func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	port, err := util.FreePort()
	if err != nil {
		t.Fatal(err)
	}
	sys := newTestSystem("http://localhost:" + strconv.Itoa(port))
	defer os.RemoveAll(sys.uiPath)

	go serve(sys, "tcp", "", port, "http")

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)
}
