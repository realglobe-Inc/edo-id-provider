package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// hndl := util.InitLog("github.com/realglobe-Inc")
	// hndl.SetLevel(level.ALL)
	// defer hndl.SetLevel(level.INFO)
	// ////////////////////////////////

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()
	routProtType := "http"

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
	go server(sys, lis, routProtType)

	// サーバ起動待ち。
	time.Sleep(100 * time.Millisecond)

	req, err := http.NewRequest("GET", "http://"+lis.Addr().String()+loginPagePath, nil)
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
