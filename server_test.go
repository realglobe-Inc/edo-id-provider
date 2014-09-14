package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// hndl := util.InitLog("github.com/realglobe-Inc")
	// hndl.SetLevel(level.ALL)
	// defer hndl.SetLevel(level.INFO)
	// ////////////////////////////////

	port, err := util.FreePort()
	if err != nil {
		t.Fatal(err)
	}

	sys := &system{
		ServiceExplorer:       driver.NewMemoryServiceExplorer(),
		ServiceKeyRegistry:    driver.NewMemoryServiceKeyRegistry(),
		UserNameIndex:         driver.NewMemoryUserNameIndex(),
		UserAttributeRegistry: driver.NewMemoryUserAttributeRegistry(),
		sessCont:              driver.NewMemoryTimeLimitedKeyValueStore(),
		codeCont:              driver.NewMemoryTimeLimitedKeyValueStore(),
		accTokenCont:          driver.NewMemoryTimeLimitedKeyValueStore(),
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
