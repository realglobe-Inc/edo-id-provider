package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
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

// edo-id-provider を立てる。
// 使い終わったら shutCh で終了させ、idpSys.uiPath を消すこと
func setupTestIdp(testAccs []*account, testTas []*ta) (idpSys *system, shutCh chan struct{}, err error) {
	port, err := util.FreePort()
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	sys := newTestSystem("http://localhost:" + strconv.Itoa(port))
	for _, acc := range testAccs {
		sys.accCont.(*memoryAccountContainer).add(acc)
	}
	for _, ta_ := range testTas {
		sys.taCont.(*memoryTaContainer).add(ta_)
	}
	shutCh = make(chan struct{}, 10)
	go serve(sys, "tcp", "", port, "http", shutCh)
	return sys, shutCh, nil
}

// 起動しただけでパニックを起こさないこと。
func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	sys, shutCh, err := setupTestIdp(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sys.uiPath)
	defer func() { shutCh <- struct{}{} }()

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)
}

// testTa を基に TA 偽装用テストサーバーを立てる。
// 使い終わったら Close すること。
func setupTestTa() (ta_ *ta, rediUri, taKid string, taPriKey crypto.PrivateKey, taServ *util.TestHttpServer, err error) {
	taPort, err := util.FreePort()
	if err != nil {
		return nil, "", "", nil, nil, erro.Wrap(err)
	}
	taServer, err := util.NewTestHttpServer(taPort)
	if err != nil {
		return nil, "", "", nil, nil, erro.Wrap(err)
	}
	taBuff := *testTa
	taBuff.Id = "http://localhost:" + strconv.Itoa(taPort)
	rediUri = taBuff.Id + "/redirect_endpoint"
	taBuff.RediUris = map[string]bool{rediUri: true}
	return &taBuff, rediUri, testTaKid, testTaPriKey, taServer, nil
}

// TA 偽装サーバーと edo-id-provider を立てる。
func setupTestTaAndIdp(testAccs []*account, testTas []*ta) (ta_ *ta, rediUri,
	taKid string, taPriKey crypto.PrivateKey, taServ *util.TestHttpServer,
	idpSys *system, shutCh chan struct{}, err error) {

	// TA 偽装サーバー。
	ta_, rediUri, taKid, taPriKey, taServ, err = setupTestTa()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			taServ.Close()
		}
	}()

	// edo-id-provider を用意。
	idpSys, shutCh, err = setupTestIdp([]*account{testAcc}, append([]*ta{ta_}, testTas...))
	return
}

// 認証リクエストを出す。
// 返り値を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testRequestAuth(idpSys *system, cli *http.Client, authParams map[string]string) (*http.Response, error) {
	q := url.Values{}
	for k, v := range authParams {
		if v != "" {
			q.Set(k, v)
		}
	}

	req, err := http.NewRequest("GET", idpSys.selfId+"/auth?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// アカウント選択 UI にリダイレクトされてたらアカウント選択する。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testSelectAccount(idpSys *system, cli *http.Client, authResp *http.Response, selParams map[string]string) (*http.Response, error) {
	if authResp.Request.URL.Path != idpSys.uiUri+"/select.html" {
		// アカウント選択 UI にリダイレクトされてない。
		return authResp, nil
	}

	if selParams == nil {
		selParams = map[string]string{}
	}

	tic := authResp.Request.URL.Fragment
	q := url.Values{}
	for k, v := range selParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	if v, ok := selParams["ticket"]; !(ok && v == "") {
		q.Set("ticket", tic)
	}
	req, err := http.NewRequest("GET", idpSys.selfId+"/auth/select?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// ログイン UI にリダイレクトされてたらログインする。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testLogin(idpSys *system, cli *http.Client, selResp *http.Response, loginParams map[string]string) (*http.Response, error) {
	if selResp.Request.URL.Path != idpSys.uiUri+"/login.html" {
		// ログイン UI にリダイレクトされてない。
		return selResp, nil
	}

	if loginParams == nil {
		loginParams = map[string]string{}
	}

	tic := selResp.Request.URL.Fragment
	q := url.Values{}
	for k, v := range loginParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	if v, ok := loginParams["ticket"]; !(ok && v == "") {
		q.Set("ticket", tic)
	}
	req, err := http.NewRequest("GET", idpSys.selfId+"/auth/login?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// 同意 UI にリダイレクトされてたら同意する。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testConsent(idpSys *system, cli *http.Client, loginResp *http.Response, consParams map[string]string) (*http.Response, error) {
	if loginResp.Request.URL.Path != idpSys.uiUri+"/consent.html" {
		// 同意 UI にリダイレクトされてない。
	}

	if consParams == nil {
		consParams = map[string]string{}
	}

	tic := loginResp.Request.URL.Fragment
	q := url.Values{}
	for k, v := range consParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	if v, ok := consParams["ticket"]; !(ok && v == "") {
		q.Set("ticket", tic)
	}
	q.Set("ticket", tic)
	req, err := http.NewRequest("GET", idpSys.selfId+"/auth/consent?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// アクセストークンを取得する。
// 返り値は JSON を Unmarshal したもの。
// パラメータ値が nil や空文字列なら、そのパラメータを設定しない。
func testGetToken(idpSys *system, consResp *http.Response, assHeads, assClms map[string]interface{},
	reqParams map[string]string, kid string, sigKey crypto.PrivateKey) (map[string]interface{}, error) {
	if assHeads == nil {
		assHeads = map[string]interface{}{}
	}
	if assClms == nil {
		assClms = map[string]interface{}{}
	}
	if reqParams == nil {
		reqParams = map[string]string{}
	}

	cod := consResp.Request.FormValue("code")
	if cod == "" {
		util.LogRequest(level.ERR, consResp.Request, true)
		return nil, erro.New("no code")
	}

	// 認可コードを取得できた。

	// クライアント認証用データを準備。

	assJws := util.NewJws()
	for k, v := range assHeads {
		assJws.SetHeader(k, v)
	}
	for k, v := range assClms {
		assJws.SetClaim(k, v)
	}
	if _, ok := assClms["code"]; !ok {
		assJws.SetClaim("code", cod)
	}
	if err := assJws.Sign(map[string]crypto.PrivateKey{kid: sigKey}); err != nil {
		return nil, erro.Wrap(err)
	}
	assBuff, err := assJws.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}
	ass := string(assBuff)

	q := url.Values{}
	for k, v := range reqParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	if v, ok := reqParams["code"]; !(ok && v == "") {
		q.Set("code", cod)
	}
	if v, ok := reqParams["client_assertion"]; !(ok && v == "") {
		q.Set("client_assertion", ass)
	}
	req, err := http.NewRequest("POST", idpSys.selfId+"/token", strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		return nil, erro.Wrap(err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		return nil, erro.Wrap(err)
	}

	return res, nil
}

// アカウント情報を取得する。
// 返り値は JSON を Unmarshal したもの。
// パラメータ値が nil や空文字列なら、そのパラメータを設定しない。
func testGetAccountInfo(idpSys *system, tokRes map[string]interface{}, reqHeads map[string]string) (map[string]interface{}, error) {
	tok, _ := tokRes["access_token"].(string)
	if tok == "" {
		return nil, erro.New("no access token")
	}

	// アクセストークンを取得できた。

	req, err := http.NewRequest("GET", idpSys.selfId+"/userinfo", nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	for k, v := range reqHeads {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	if v, ok := reqHeads["Authorization"]; !(ok && v == "") {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return res, nil
}

// 認証リクエストからアカウント情報取得までする。
func testFromRequestAuthToGetAccountInfo(idpSys *system, cli *http.Client,
	authParams, selParams, loginParams, consParams map[string]string,
	assHeads, assClms map[string]interface{}, tokParams map[string]string, kid string, sigKey crypto.PrivateKey,
	accInfHeads map[string]string) (map[string]interface{}, error) {

	// リクエストする。
	authResp, err := testRequestAuth(idpSys, cli, authParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer authResp.Body.Close()

	// 必要ならアカウント選択する。
	selResp, err := testSelectAccount(idpSys, cli, authResp, selParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer selResp.Body.Close()

	// 必要ならログインする。
	loginResp, err := testLogin(idpSys, cli, selResp, loginParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer loginResp.Body.Close()

	// 必要なら同意する。
	consResp, err := testConsent(idpSys, cli, loginResp, consParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer consResp.Body.Close()

	// アクセストークンを取得する。
	tokRes, err := testGetToken(idpSys, consResp, assHeads, assClms, tokParams, kid, sigKey)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	// アカウント情報を取得する。
	return testGetAccountInfo(idpSys, tokRes, accInfHeads)
}

// 認証してアカウント情報を取得できるか。
func TestSuccess(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	// 認可コードのリダイレクト先としての TA を用意。
	testTa2, rediUri, kid, sigKey, taServ, err := setupTestTa()
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	// 無事 TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// edo-id-provider を用意。
	sys, shutCh, err := setupTestIdp([]*account{testAcc}, []*ta{testTa2})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sys.uiPath)
	defer func() { shutCh <- struct{}{} }()

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	res, err := testFromRequestAuthToGetAccountInfo(sys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"prompt":        "select_account login consent",
	}, map[string]string{
		"username": testAcc.name(),
	}, map[string]string{
		"username": testAcc.name(),
		"password": testAcc.password(),
	}, map[string]string{
		"consented_scope": "openid email",
	}, map[string]interface{}{
		"alg": "RS256",
		"kid": kid,
	}, map[string]interface{}{
		"iss": testTa2.id(),
		"sub": testTa2.id(),
		"aud": sys.selfId + "/token",
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(sys.idTokExpiDur).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          rediUri,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey, nil)
	if err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != testAcc.attribute("email") {
		t.Fatal(em, testAcc.attribute("email"))
	}
}
