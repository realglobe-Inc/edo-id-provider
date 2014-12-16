package main

import (
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"os"
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
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	sys := &system{
		uiUri:    "/login/html",
		uiPath:   path,
		taCont:   newMemoryTaContainer(0, 0),
		accCont:  newMemoryAccountContainer(0, 0),
		sessCont: newMemorySessionContainer(10, time.Second, 0, 0),
		codCont:  newMemoryCodeContainer(10, time.Second, 0, 0),
		tokCont:  newMemoryTokenContainer(10, time.Second, time.Second, 0, 0),
	}
	go serve(sys, "tcp", "", port, "http")

	// サーバ起動待ち。
	time.Sleep(50 * time.Millisecond)

	// req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port)+loginPagePath, nil)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// resp, err := (&http.Client{}).Do(req)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	t.Error(resp)
	// }
}
