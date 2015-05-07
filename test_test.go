// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/test"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// テスト用の ID プロバイダ。
type testIdpServer struct {
	dir string
	sys *system
	*httptest.Server
	stopper *server.Stopper
}

// テスト用の ID プロバイダを立てる。
func newTestIdpServer(acnts []account.Element,
	tas []tadb.Element,
	idps []idpdb.Element,
	webs []webdb.Element) (*testIdpServer, error) {

	failed := true

	// UI 用 HTML を用意。
	dir, err := ioutil.TempDir("", test_label)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer func() {
		if failed {
			os.RemoveAll(dir)
		}
	}()

	html := []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>ダミー</title></head><body><p>ダミー</p></body></html>`)
	if err := ioutil.WriteFile(filepath.Join(dir, strings.TrimPrefix(test_pathSelUi, test_pathUi)), html, 0644); err != nil {
		return nil, erro.Wrap(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, strings.TrimPrefix(test_pathLginUi, test_pathUi)), html, 0644); err != nil {
		return nil, erro.Wrap(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, strings.TrimPrefix(test_pathConsUi, test_pathUi)), html, 0644); err != nil {
		return nil, erro.Wrap(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, strings.TrimPrefix(test_pathErrUi, test_pathUi)), html, 0644); err != nil {
		return nil, erro.Wrap(err)
	}

	// system を用意。
	sys := newTestSystem([]jwk.Key{test_idpPriKey}, acnts, tas, idps, webs)

	// サーバーを設定。
	s := server.NewStopper()

	mux := http.NewServeMux()
	mux.HandleFunc(test_pathOk, panicErrorWrapper(s, func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}))
	mux.HandleFunc(test_pathAuth, panicErrorWrapper(s, sys.authPage))
	mux.HandleFunc(test_pathSel, panicErrorWrapper(s, sys.selectPage))
	mux.HandleFunc(test_pathLgin, panicErrorWrapper(s, sys.lginPage))
	mux.HandleFunc(test_pathCons, panicErrorWrapper(s, sys.consentPage))
	mux.HandleFunc(test_pathTa, panicErrorWrapper(s, sys.taApiHandler().ServeHTTP))
	mux.HandleFunc(test_pathTok, panicErrorWrapper(s, sys.tokenApi))
	mux.HandleFunc(test_pathAcnt, panicErrorWrapper(s, sys.accountApi))
	mux.HandleFunc(test_pathCoopFr, panicErrorWrapper(s, sys.cooperateFromApi))
	mux.HandleFunc(test_pathCoopTo, panicErrorWrapper(s, sys.cooperateToApi))
	filer := http.StripPrefix(test_pathUi+"/", http.FileServer(http.Dir(dir)))
	mux.Handle(test_pathUi+"/", filer)

	server := httptest.NewServer(mux)
	defer func() {
		if failed {
			server.Close()
		}
	}()

	// これで同期されるかどうかは不明。
	sys.selfId = server.URL
	// 念のため叩いてみるけど、これでも同期されるかどうかは不明。
	if _, err := http.Get(server.URL + test_pathOk); err != nil {
		return nil, erro.Wrap(err)
	}

	failed = false
	return &testIdpServer{dir, sys, server, s}, nil
}

func (this *testIdpServer) close() {
	this.Server.Close()
	this.stopper.Lock()
	defer this.stopper.Unlock()
	for this.stopper.Stopped() {
		this.stopper.Wait()
	}
	os.RemoveAll(this.dir)
}

// ダミー TA。
type testTaServer struct {
	info tadb.Element
	*test.HttpServer
	rediUri string
}

// ダミー TA を立てる。
func newTestTaServer(rediPath string) (*testTaServer, error) {
	failed := true
	server, err := test.NewHttpServer(30 * time.Second)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer func() {
		if failed {
			server.Close()
		}
	}()

	rediUris := map[string]bool{server.URL + test_taPathCb: true}
	rediUri := server.URL
	if rediPath == "" {
		rediUri += test_taPathCb
	} else {
		rediUri += rediPath
		rediUris[rediUri] = true
	}
	info := tadb.New(server.URL,
		map[string]string{"": test_taName, "ja": test_taNameJa},
		rediUris,
		[]jwk.Key{test_taPubKey},
		true,
		server.URL)

	failed = false
	return &testTaServer{info, server, rediUri}, nil
}

func (this *testTaServer) close() {
	this.HttpServer.Close()
}

func (this *testTaServer) taInfo() tadb.Element {
	return this.info
}

func (this *testTaServer) redirectUri() string {
	return this.rediUri
}

func (this *testTaServer) webInfo() []webdb.Element {
	return nil
}

// テスト用 ID プロバイダとダミー TA を立てる。
func setupTestIdpAndTa(acnts []account.Element,
	tas []tadb.Element,
	idps []idpdb.Element,
	webs []webdb.Element) (*testIdpServer, *testTaServer, error) {

	failed := true

	ta, err := newTestTaServer("")
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	defer func() {
		if failed {
			ta.close()
		}
	}()

	if tas == nil {
		tas = []tadb.Element{}
	}
	tas = append(tas, ta.taInfo())
	if webs == nil {
		webs = []webdb.Element{}
	}
	webs = append(webs, ta.webInfo()...)

	idp, err := newTestIdpServer(acnts, tas, idps, webs)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	failed = false
	return idp, ta, nil
}

// テスト用のアカウントをつくる。
func newTestAccount() account.Element {
	return account.New(test_acntId, test_acntName, test_acntAuth, clone(test_acntAttrs))
}

// 1 段目だけのコピー。
func clone(m map[string]interface{}) map[string]interface{} {
	m2 := map[string]interface{}{}
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// 認証リクエストを出し結果を無検査で返す。
// 返り値を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testRequestAuthWithoutCheck(idp *testIdpServer, cli *http.Client, authParams map[string]string) (*http.Response, error) {
	q := url.Values{}
	for k, v := range authParams {
		if v != "" {
			q.Set(k, v)
		}
	}

	req, err := http.NewRequest("GET", idp.sys.selfId+test_pathAuth+"?"+q.Encode(), nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Connection", "close")
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return resp, nil
}

// 認証リクエストを出す。
// 返り値を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testRequestAuth(idp *testIdpServer, cli *http.Client, authParams map[string]string) (*http.Response, error) {
	resp, err := testRequestAuthWithoutCheck(idp, cli, authParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// アカウント選択 UI にリダイレクトされてたらアカウント選択して結果を無検査で返す。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testSelectAccountWithoutCheck(idp *testIdpServer, cli *http.Client, authResp *http.Response, selParams map[string]string) (*http.Response, error) {
	if authResp.Request.URL.Path != test_pathSelUi {
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
	if _, ok := selParams["ticket"]; !ok {
		q.Set("ticket", tic)
	}
	req, err := http.NewRequest("POST", idp.sys.selfId+test_pathSel, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", server.ContentTypeForm)
	req.Header.Set("Connection", "close")
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return resp, nil
}

// アカウント選択 UI にリダイレクトされてたらアカウント選択する。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testSelectAccount(idp *testIdpServer, cli *http.Client, authResp *http.Response, selParams map[string]string) (*http.Response, error) {
	resp, err := testSelectAccountWithoutCheck(idp, cli, authResp, selParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// ログイン UI にリダイレクトされてたらログインして結果を無検査で返す。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testLoginWithoutCheck(idp *testIdpServer, cli *http.Client, selResp *http.Response, lginParams map[string]string) (*http.Response, error) {
	if selResp.Request.URL.Path != test_pathLginUi {
		// ログイン UI にリダイレクトされてない。
		return selResp, nil
	}

	if lginParams == nil {
		lginParams = map[string]string{}
	}

	tic := selResp.Request.URL.Fragment
	q := url.Values{}
	for k, v := range lginParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	if _, ok := lginParams["ticket"]; !ok {
		q.Set("ticket", tic)
	}
	req, err := http.NewRequest("POST", idp.sys.selfId+test_pathLgin, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", server.ContentTypeForm)
	req.Header.Set("Connection", "close")
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return resp, nil
}

// ログイン UI にリダイレクトされてたらログインする。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testLogin(idp *testIdpServer, cli *http.Client, selResp *http.Response, lginParams map[string]string) (*http.Response, error) {
	resp, err := testLoginWithoutCheck(idp, cli, selResp, lginParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// 同意 UI にリダイレクトされてたら同意して結果を無検査で返す。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testConsentWithoutCheck(idp *testIdpServer, cli *http.Client, lginResp *http.Response, consParams map[string]string) (*http.Response, error) {
	if lginResp.Request.URL.Path != test_pathConsUi {
		// 同意 UI にリダイレクトされてない。
		return lginResp, nil
	}

	if consParams == nil {
		consParams = map[string]string{}
	}

	tic := lginResp.Request.URL.Fragment
	q := url.Values{}
	for k, v := range consParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	if _, ok := consParams["ticket"]; !ok {
		q.Set("ticket", tic)
	}
	q.Set("ticket", tic)
	req, err := http.NewRequest("POST", idp.sys.selfId+test_pathCons, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", server.ContentTypeForm)
	req.Header.Set("Connection", "close")
	resp, err := cli.Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return resp, nil
}

// 同意 UI にリダイレクトされてたら同意する。
// 返り値の Body を Close すること。
// パラメータ値が空文字列なら、そのパラメータを設定しない。
func testConsent(idp *testIdpServer, cli *http.Client, lginResp *http.Response, consParams map[string]string) (*http.Response, error) {
	resp, err := testConsentWithoutCheck(idp, cli, lginResp, consParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}
	return resp, nil
}

// トークンリクエストして結果を無検査で返す。
// 返り値の Body を Close すること。
// パラメータ値が空値なら、そのパラメータを設定しない。
func testGetTokenWithoutCheck(idp *testIdpServer, consResp *http.Response, assHeads, assClms map[string]interface{},
	reqParams map[string]string, sigKey jwk.Key) (*http.Response, error) {
	cod := consResp.Request.FormValue("code")
	if cod == "" {
		server.LogRequest(level.ERR, consResp.Request, true)
		return nil, erro.New("no code")
	}

	// 認可コードを取得できた。

	if assHeads == nil {
		assHeads = map[string]interface{}{}
	}
	if assClms == nil {
		assClms = map[string]interface{}{}
	}
	if reqParams == nil {
		reqParams = map[string]string{}
	}

	// クライアント認証用データを準備。

	assJt := jwt.New()
	for k, v := range assHeads {
		assJt.SetHeader(k, v)
	}
	for k, v := range assClms {
		assJt.SetClaim(k, v)
	}
	if _, ok := assClms["code"]; !ok {
		assJt.SetClaim("code", cod)
	}
	if err := assJt.Sign([]jwk.Key{sigKey}); err != nil {
		return nil, erro.Wrap(err)
	}
	assBuff, err := assJt.Encode()
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
	if _, ok := reqParams["code"]; !ok {
		q.Set("code", cod)
	}
	if _, ok := reqParams["client_assertion"]; !ok {
		q.Set("client_assertion", ass)
	}
	req, err := http.NewRequest("POST", idp.sys.selfId+test_pathTok, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	req.Header.Set("Content-Type", server.ContentTypeForm)
	req.Header.Set("Connection", "close")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return resp, nil
}

// アクセストークンを取得する。
// 返り値は JSON を Unmarshal したもの。
// パラメータ値が空値なら、そのパラメータを設定しない。
func testGetToken(idp *testIdpServer, consResp *http.Response, assHeads, assClms map[string]interface{},
	reqParams map[string]string, sigKey jwk.Key) (map[string]interface{}, error) {
	resp, err := testGetTokenWithoutCheck(idp, consResp, assHeads, assClms, reqParams, sigKey)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}

	var res map[string]interface{}
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		return nil, erro.Wrap(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		return nil, erro.Wrap(err)
	}
	return res, nil
}

// アカウント情報を取得する。
// 返り値の Body を Close すること。
// パラメータ値が空値なら、そのパラメータを設定しない。
func testGetAccountInfoWithoutCheck(idp *testIdpServer, tokRes map[string]interface{}, reqHeads map[string]string) (*http.Response, error) {
	tok, _ := tokRes["access_token"].(string)
	if tok == "" {
		return nil, erro.New("no access token")
	}

	// アクセストークンを取得できた。

	req, err := http.NewRequest("GET", idp.sys.selfId+test_pathAcnt, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	for k, v := range reqHeads {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	if _, ok := reqHeads["Authorization"]; !ok {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	req.Header.Set("Connection", "close")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return resp, nil
}

// アカウント情報を取得する。
// 返り値は JSON を Unmarshal したもの。
// パラメータ値が空値なら、そのパラメータを設定しない。
func testGetAccountInfo(idp *testIdpServer, tokRes map[string]interface{}, reqHeads map[string]string) (map[string]interface{}, error) {
	resp, err := testGetAccountInfoWithoutCheck(idp, tokRes, reqHeads)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		resp.Body.Close()
		return nil, erro.New("invalid response ", resp.StatusCode, " "+http.StatusText(resp.StatusCode))
	}

	var res map[string]interface{}
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, erro.Wrap(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return res, nil
}

// 認証リクエストから認可コード取得までする。
func testFromRequestAuthToConsent(idp *testIdpServer, cli *http.Client,
	authParams, selParams, lginParams, consParams map[string]string) (*http.Response, error) {

	// リクエストする。
	authResp, err := testRequestAuth(idp, cli, authParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer authResp.Body.Close()

	// 必要ならアカウント選択する。
	selResp, err := testSelectAccount(idp, cli, authResp, selParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer selResp.Body.Close()

	// 必要ならログインする。
	lginResp, err := testLogin(idp, cli, selResp, lginParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer lginResp.Body.Close()

	// 必要なら同意する。
	return testConsent(idp, cli, lginResp, consParams)
}

// 認証リクエストからトークンリクエストまでして結果を無検査で返す。
func testFromRequestAuthToGetTokenWithoutCheck(idp *testIdpServer, cli *http.Client,
	authParams, selParams, lginParams, consParams map[string]string,
	assHeads, assClms map[string]interface{}, tokParams map[string]string, sigKey jwk.Key) (*http.Response, error) {

	// リクエストから同意までする。
	consResp, err := testFromRequestAuthToConsent(idp, cli, authParams, selParams, lginParams, consParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer consResp.Body.Close()

	// アクセストークンを取得してアカウント情報を取得する。
	return testGetTokenWithoutCheck(idp, consResp, assHeads, assClms, tokParams, sigKey)
}

// 認証リクエストからトークンリクエストまでする。
func testFromRequestAuthToGetToken(idp *testIdpServer, cli *http.Client,
	authParams, selParams, lginParams, consParams map[string]string,
	assHeads, assClms map[string]interface{}, tokParams map[string]string, sigKey jwk.Key) (map[string]interface{}, error) {

	// リクエストから同意までする。
	consResp, err := testFromRequestAuthToConsent(idp, cli, authParams, selParams, lginParams, consParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer consResp.Body.Close()

	// アクセストークンを取得してアカウント情報を取得する。
	return testGetToken(idp, consResp, assHeads, assClms, tokParams, sigKey)
}

// トークンリクエストからアカウント情報取得までする。
func testGetTokenAndAccountInfo(idp *testIdpServer, consResp *http.Response,
	assHeads, assClms map[string]interface{}, tokParams map[string]string, sigKey jwk.Key,
	accInfHeads map[string]string) (map[string]interface{}, error) {

	// アクセストークンを取得する。
	tokRes, err := testGetToken(idp, consResp, assHeads, assClms, tokParams, sigKey)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	// アカウント情報を取得する。
	return testGetAccountInfo(idp, tokRes, accInfHeads)
}

// 認証リクエストからアカウント情報取得までする。
func testFromRequestAuthToGetAccountInfo(idp *testIdpServer, cli *http.Client,
	authParams, selParams, lginParams, consParams map[string]string,
	assHeads, assClms map[string]interface{}, tokParams map[string]string, sigKey jwk.Key,
	accInfHeads map[string]string) (map[string]interface{}, error) {

	// リクエストから同意までする。
	consResp, err := testFromRequestAuthToConsent(idp, cli, authParams, selParams, lginParams, consParams)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer consResp.Body.Close()

	// アクセストークンを取得してアカウント情報を取得する。
	return testGetTokenAndAccountInfo(idp, consResp, assHeads, assClms, tokParams, sigKey, accInfHeads)
}
