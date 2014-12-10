package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func init() {
	util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
}

func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	port, err := util.FreePort()
	if err != nil {
		t.Fatal(err)
	}

	sys := &system{
		TaExplorer:            NewMemoryTaExplorer(0),
		TaKeyProvider:         NewMemoryTaKeyProvider(0),
		UserNameIndex:         NewMemoryUserNameIndex(0),
		UserAttributeRegistry: NewMemoryUserAttributeRegistry(0),
		sessCont:              driver.NewMemoryTimeLimitedKeyValueStore(0),
		codeCont:              driver.NewMemoryTimeLimitedKeyValueStore(0),
		accTokenCont:          driver.NewMemoryTimeLimitedKeyValueStore(0),
		maxSessExpiDur:        time.Hour,
		codeExpiDur:           time.Hour,
		accTokenExpiDur:       time.Hour,
		maxAccTokenExpiDur:    time.Hour,
	}
	go serve(sys, "tcp", "", port, "http")

	// サーバ起動待ち。
	time.Sleep(50 * time.Millisecond)

	req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port)+loginPagePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Error(resp)
	}
}
