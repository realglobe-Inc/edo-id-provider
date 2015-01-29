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
	idpSys = newTestSystem("http://localhost:" + strconv.Itoa(port))
	for _, acc := range testAccs {
		idpSys.accCont.(*memoryAccountContainer).add(acc)
	}
	for _, ta_ := range testTas {
		idpSys.taCont.(*memoryTaContainer).add(ta_)
	}
	shutCh = make(chan struct{}, 10)
	go serve(idpSys, "tcp", "", port, "http", shutCh)
	return idpSys, shutCh, nil
}

// 起動しただけでパニックを起こさないこと。
func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idpSys, shutCh, err := setupTestIdp(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)
}

// testTa を基に TA 偽装用テストサーバーを立てる。
// 使い終わったら Close すること。
func setupTestTa(rediUriPaths []string) (ta_ *ta, rediUri, taKid string, taPriKey crypto.PrivateKey, taServ *util.TestHttpServer, err error) {
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
	if len(rediUriPaths) == 0 {
		rediUri = taBuff.Id + "/redirect_endpoint"
		taBuff.RediUris = map[string]bool{rediUri: true}
	} else {
		taBuff.RediUris = map[string]bool{}
		for _, v := range rediUriPaths {
			rediUri = taBuff.Id + v
			taBuff.RediUris[rediUri] = true
		}
	}
	return &taBuff, rediUri, testTaKid, testTaPriKey, taServer, nil
}

// TA 偽装サーバーと edo-id-provider を立てる。
func setupTestTaAndIdp(rediUriPaths []string, testAccs []*account, testTas []*ta) (ta_ *ta, rediUri,
	taKid string, taPriKey crypto.PrivateKey, taServ *util.TestHttpServer,
	idpSys *system, shutCh chan struct{}, err error) {

	// TA 偽装サーバー。
	ta_, rediUri, taKid, taPriKey, taServ, err = setupTestTa(rediUriPaths)
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
	req, err := http.NewRequest("POST", idpSys.selfId+"/auth/select", strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", util.ContentTypeForm)
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
	req, err := http.NewRequest("POST", idpSys.selfId+"/auth/login", strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", util.ContentTypeForm)
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
		return loginResp, nil
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
	req, err := http.NewRequest("POST", idpSys.selfId+"/auth/consent", strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", util.ContentTypeForm)
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
	req.Header.Set("Content-Type", util.ContentTypeForm)
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

// 認証リクエストから認可コード取得までする。
func testFromRequestAuthToConsent(idpSys *system, cli *http.Client,
	authParams, selParams, loginParams, consParams map[string]string) (*http.Response, error) {

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
	return testConsent(idpSys, cli, loginResp, consParams)
}

// 認証リクエストからアカウント情報取得までする。
func testFromRequestAuthToGetAccountInfo(idpSys *system, cli *http.Client,
	authParams, selParams, loginParams, consParams map[string]string,
	assHeads, assClms map[string]interface{}, tokParams map[string]string, kid string, sigKey crypto.PrivateKey,
	accInfHeads map[string]string) (map[string]interface{}, error) {

	// リクエストから同意までする。
	consResp, err := testFromRequestAuthToConsent(idpSys, cli, authParams, selParams, loginParams, consParams)
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

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	res, err := testFromRequestAuthToGetAccountInfo(idpSys, cli, map[string]string{
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
		"aud": idpSys.selfId + "/token",
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idpSys.idTokExpiDur).Unix(),
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

// 知らないパラメータを無視できるか。
func TestIgnoreUnknownParameterInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	res, err := testFromRequestAuthToGetAccountInfo(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"unknown_name":  "unknown_value",
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
		"iss":     testTa2.id(),
		"sub":     testTa2.id(),
		"aud":     idpSys.selfId + "/token",
		"jti":     strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp":     time.Now().Add(idpSys.idTokExpiDur).Unix(),
		"unknown": "unknown",
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          rediUri,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		"unknown":               "unknown",
	}, kid, sigKey, nil)
	if err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != testAcc.attribute("email") {
		t.Fatal(em, testAcc.attribute("email"))
	}
}

// 認証リクエストの重複パラメータを拒否できるか。
func TestDenyOverlapParameterInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	q := url.Values{
		"scope":         {"openid email"},
		"response_type": {"code"},
		"client_id":     {testTa2.id()},
		"redirect_uri":  {rediUri},
	}

	req, err := http.NewRequest("GET", idpSys.selfId+"/auth?"+q.Encode()+"&scope=aaaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusOK)
	} else if resp.Request.FormValue(formErr) == "" {
		t.Fatal("no error")
	}
}

// 認証リクエストに response_type が無い時に拒否できるか。
func TestDenyNoResponseTypeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	resp, err := testRequestAuth(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Request.FormValue(formErr) != errInvReq {
		t.Fatal(resp.Request.FormValue(formErr), errInvReq)
	}
}

// 認証リクエストの response_type が未知の時に拒否できるか。
func TestDenyUnknownResponseTypeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	resp, err := testRequestAuth(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "unknown",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Request.FormValue(formErr) != errUnsuppRespType {
		t.Fatal(resp.Request.FormValue(formErr), errInvReq)
	}
}

// 結果をリダイレクトで返すときに redirect_uri のパラメータを維持できるか。
func TestKeepRedirectUriParameter(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/redirect_endpoint?param_name=param_value"}, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	resp, err := testFromRequestAuthToConsent(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	}, map[string]string{
		"username": testAcc.name(),
	}, map[string]string{
		"username": testAcc.name(),
		"password": testAcc.password(),
	}, map[string]string{
		"consented_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	} else if q.Get("param_name") != "param_value" {
		t.Fatal(q.Get("param_name"), "param_value")
	}
}

// エラーをリダイレクトで返すときに redirect_uri のパラメータを維持できるか。
func TestKeepRedirectUriParameterInError(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/redirect_endpoint?param_name=param_value"}, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	resp, err := testRequestAuth(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "unknown",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("error") != errUnsuppRespType {
		t.Fatal(q.Get("error"), errUnsuppRespType)
	} else if q.Get("param_name") != "param_value" {
		t.Fatal(q.Get("param_name"), "param_value")
	}
}

// redirect_uri が登録値と異なるときにリダイレクトせずに拒否できるか。
func TestDirectErrorResponseInInvalidRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	req, err := http.NewRequest("GET", idpSys.selfId+"/auth?"+url.Values{
		"scope":         {"openid email"},
		"response_type": {"code"},
		"client_id":     {testTa2.id()},
		"redirect_uri":  {rediUri + "/"},
	}.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	}

	var res struct{ Error string }
	if err := json.Unmarshal(data, &res); err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// redirect_uri が無いときにリダイレクトせずに拒否できるか。
func TestDirectErrorResponseInNoRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, _, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	req, err := http.NewRequest("GET", idpSys.selfId+"/auth?"+url.Values{
		"scope":         {"openid email"},
		"response_type": {"code"},
		"client_id":     {testTa2.id()},
	}.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	}

	var res struct{ Error string }
	if err := json.Unmarshal(data, &res); err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// POST でないトークンリクエストを拒否できるか。
func TestDenyNonPostTokenRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	for _, meth := range []string{"GET", "PUT"} {
		// TA にリダイレクトできたときのレスポンスを設定しておく。
		taServ.AddResponse(http.StatusOK, nil, []byte("success"))

		cookJar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		cli := &http.Client{Jar: cookJar}

		consResp, err := testFromRequestAuthToConsent(idpSys, cli, map[string]string{
			"scope":         "openid email",
			"response_type": "code",
			"client_id":     testTa2.id(),
			"redirect_uri":  rediUri,
		}, map[string]string{
			"username": testAcc.name(),
		}, map[string]string{
			"username": testAcc.name(),
			"password": testAcc.password(),
		}, map[string]string{
			"consented_scope": "openid email",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer consResp.Body.Close()

		cod := consResp.Request.FormValue("code")
		if cod == "" {
			util.LogRequest(level.ERR, consResp.Request, true)
			t.Fatal("no code")
		}

		// 認可コードを取得できた。

		assJws := util.NewJws()
		assJws.SetHeader("alg", "RS256")
		assJws.SetHeader("kid", kid)
		assJws.SetClaim("iss", testTa2.id())
		assJws.SetClaim("sub", testTa2.id())
		assJws.SetClaim("aud", idpSys.selfId+"/token")
		assJws.SetClaim("jti", strconv.FormatInt(time.Now().UnixNano(), 16))
		assJws.SetClaim("exp", time.Now().Add(idpSys.idTokExpiDur).Unix())
		assJws.SetClaim("code", cod)
		if err := assJws.Sign(map[string]crypto.PrivateKey{kid: sigKey}); err != nil {
			t.Fatal(err)
		}
		assBuff, err := assJws.Encode()
		if err != nil {
			t.Fatal(err)
		}
		ass := string(assBuff)

		req, err := http.NewRequest(meth, idpSys.selfId+"/token", strings.NewReader(url.Values{
			"grant_type":            {"authorization_code"},
			"redirect_uri":          {rediUri},
			"client_id":             {testTa2.id()},
			"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
			"code":                  {cod},
			"client_assertion":      {ass},
		}.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			util.LogRequest(level.ERR, req, true)
			util.LogResponse(level.ERR, resp, true)
			t.Fatal(resp.StatusCode, http.StatusMethodNotAllowed)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			util.LogRequest(level.ERR, req, true)
			util.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		}

		var res struct{ Error string }
		if err := json.Unmarshal(data, &res); err != nil {
			util.LogRequest(level.ERR, req, true)
			util.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if res.Error != errInvReq {
			t.Fatal(res.Error, errInvReq)
		}
	}
}

// トークンリクエストの未知のパラメータを無視できるか。
func TestIgnoreUnknownParameterInTokenRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	res, err := testFromRequestAuthToGetAccountInfo(idpSys, cli, map[string]string{
		"scope":         "openid",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
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
		"aud": idpSys.selfId + "/token",
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idpSys.idTokExpiDur).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          rediUri,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		"unknown_name":          "unknown_value",
	}, kid, sigKey, nil)
	if err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != testAcc.attribute("email") {
		t.Fatal(em, testAcc.attribute("email"))
	}
}

// 認証中にエラーが起きたら認証経過を破棄できるか。
func TestAbortSession(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	// リクエストする。
	authResp, err := testRequestAuth(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"prompt":        "select_account",
		"unknown":       "unknown",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer authResp.Body.Close()

	// アカウント選択でアカウント選択券を渡さないで認証経過をリセット。
	selResp, err := testSelectAccount(idpSys, cli, authResp, map[string]string{
		"username": testAcc.name(),
		"ticket":   "",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer selResp.Body.Close()

	if selResp.Request.FormValue(formErr) != errAccDeny {
		t.Fatal(selResp.Request.FormValue(formErr), errAccDeny)
	}

	// アカウント選択でさっきのアカウント選択券を渡す。
	q := url.Values{
		"username": {testAcc.name()},
		"ticket":   {authResp.Request.URL.Fragment},
	}
	req, err := http.NewRequest("GET", idpSys.selfId+"/auth/select?"+q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	}

	var res struct{ Error string }
	if err := json.Unmarshal(data, &res); err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}
