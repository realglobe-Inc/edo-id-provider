package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	go serve(sys, "tcp", "", port, "http", nil)

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)
}

// 認証してアカウント情報を取得できるか。
func TestSuccess(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	// 認可コードのリダイレクト先としての TA を用意。
	taPort, err := util.FreePort()
	if err != nil {
		t.Fatal(err)
	}
	taServer, err := util.NewTestHttpServer(taPort)
	if err != nil {
		t.Fatal(err)
	}
	defer taServer.Close()
	taBuff := *testTa
	taBuff.Id = "http://localhost:" + strconv.Itoa(taPort)
	taBuff.RediUris = map[string]bool{taBuff.Id + "/redirect_endpoint": true}
	testTa2 := &taBuff

	// edo-id-provider を用意。
	port, err := util.FreePort()
	if err != nil {
		t.Fatal(err)
	}
	sys := newTestSystem("http://localhost:" + strconv.Itoa(port))
	defer os.RemoveAll(sys.uiPath)
	sys.accCont.(*memoryAccountContainer).add(testAcc)
	sys.taCont.(*memoryTaContainer).add(testTa2)
	go serve(sys, "tcp", "", port, "http", nil)

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	// 無事 TA にリダイレクトできたときのレスポンスを設定しておく。
	taServer.AddResponse(http.StatusOK, nil, []byte("success"))

	rediUri := util.OneOfStringSet(testTa2.redirectUris())
	q := url.Values{}
	q.Set("scope", "openid email")
	q.Set("response_type", "code")
	q.Set("client_id", testTa2.id())
	q.Set("redirect_uri", rediUri)
	q.Set("prompt", "select_account login consent")
	req, err := http.NewRequest("GET", sys.selfId+"/auth?"+q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	cookJar, err := cookiejar.New(nil)
	resp, err := (&http.Client{Jar: cookJar}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buff, _ := httputil.DumpRequest(req, false)
		t.Error(string(buff))
		buff, _ = httputil.DumpResponse(resp, true)
		t.Fatal(string(buff))
	}

	// アカウント選択が必要。
	if resp.Request.URL.Path == sys.uiUri+"/select.html" {
		tic := resp.Request.URL.Fragment

		q := url.Values{}
		q.Set("username", testAcc.name())
		q.Set("ticket", tic)
		req, err = http.NewRequest("GET", sys.selfId+"/auth/select?"+q.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err = (&http.Client{Jar: cookJar}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			buff, _ := httputil.DumpRequest(req, false)
			t.Error(string(buff))
			buff, _ = httputil.DumpResponse(resp, true)
			t.Fatal(string(buff))
		}
	}

	// ログインが必要。
	if resp.Request.URL.Path == sys.uiUri+"/login.html" {
		tic := resp.Request.URL.Fragment

		q := url.Values{}
		q.Set("username", testAcc.name())
		q.Set("password", testAcc.password())
		q.Set("ticket", tic)
		req, err = http.NewRequest("GET", sys.selfId+"/auth/login?"+q.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err = (&http.Client{Jar: cookJar}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			buff, _ := httputil.DumpRequest(req, false)
			t.Error(string(buff))
			buff, _ = httputil.DumpResponse(resp, true)
			t.Fatal(string(buff))
		}
	}

	// 同意が必要。
	if resp.Request.URL.Path == sys.uiUri+"/consent.html" {
		tic := resp.Request.URL.Fragment

		q := url.Values{}
		q.Set("consented_scope", "openid email")
		q.Set("ticket", tic)
		req, err = http.NewRequest("GET", sys.selfId+"/auth/consent?"+q.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err = (&http.Client{Jar: cookJar}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			buff, _ := httputil.DumpRequest(req, false)
			t.Error(string(buff))
			buff, _ = httputil.DumpResponse(resp, true)
			t.Fatal(string(buff))
		}
	}

	cod := resp.Request.FormValue("code")
	if cod == "" {
		buff, _ := httputil.DumpRequest(req, false)
		t.Error(string(buff))
		buff, _ = httputil.DumpResponse(resp, true)
		t.Fatal(string(buff))
	}

	// 認可コードを取得できた。

	// クライアント認証用データを準備。
	assJws := util.NewJws()
	assJws.SetHeader("alg", "RS256")
	assJws.SetClaim("iss", testTa2.id())
	assJws.SetClaim("sub", testTa2.id())
	assJws.SetClaim("aud", sys.selfId+"/token")
	assJws.SetClaim("jti", "abcde")
	assJws.SetClaim("exp", time.Now().Add(sys.idTokExpiDur).Unix())
	assJws.SetClaim("cod", cod)
	if err := assJws.Sign(map[string]crypto.PrivateKey{"": testTaPriKey}); err != nil {
		t.Fatal(err)
	}
	assBuff, err := assJws.Encode()
	if err != nil {
		t.Fatal(err)
	}
	ass := string(assBuff)

	q = url.Values{}
	q.Set("grant_type", "authorization_code")
	q.Set("code", cod)
	q.Set("redirect_uri", rediUri)
	q.Set("client_id", testTa2.id())
	q.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	q.Set("client_assertion", ass)
	req, err = http.NewRequest("POST", sys.selfId+"/token", strings.NewReader(q.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var tokRes struct {
		Access_token string
	}
	if err := json.Unmarshal(data, &tokRes); err != nil {
		t.Fatal(err)
	} else if tokRes.Access_token == "" {
		buff, _ := httputil.DumpRequest(req, true)
		t.Error(string(buff))
		buff, _ = httputil.DumpResponse(resp, true)
		t.Fatal(string(buff))
	}

	// アクセストークンを取得できた。

	req, err = http.NewRequest("GET", sys.selfId+"/userinfo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokRes.Access_token)
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var usrInfRes struct {
		Email string
	}
	if err := json.Unmarshal(data, &usrInfRes); err != nil {
		t.Fatal(err)
	} else if usrInfRes.Email != testAcc.attribute("email") {
		buff, _ := httputil.DumpRequest(req, true)
		t.Error(string(buff))
		buff, _ = httputil.DumpResponse(resp, true)
		t.Fatal(string(buff))
	}
}
