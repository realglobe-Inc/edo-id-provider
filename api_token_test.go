package main

// トークンリクエスト・レスポンス周りのテスト。

import (
	"crypto"
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
}

// トークンレスポンスが access_token, token_type, expires_in を含むか。
func TestTokenResponse(t *testing.T) {
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

	if res, err := testFromRequestAuthToGetToken(idpSys, cli, map[string]string{
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
	}, kid, sigKey); err != nil {
		t.Fatal(err)
	} else if tok, _ := res["access_token"].(string); tok == "" {
		t.Fatal("no token")
	} else if respType, _ := res["token_type"].(string); respType == "" {
		t.Fatal("no token type")
	} else if _, ok := res["expires_in"].(float64); !ok {
		t.Fatal("no expiration date")
	}
}

// 要求された scope と異なるトークンを返すとき scope を含むか。
func TestTokenResponseWithScope(t *testing.T) {
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

	if res, err := testFromRequestAuthToGetToken(idpSys, cli, map[string]string{
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
		"consented_scope": "openid",
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
	}, kid, sigKey); err != nil {
		t.Fatal(err)
	} else if scop, _ := res["scope"].(string); scop == "" {
		t.Fatal("no scope")
	}
}

// トークンレスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むか。
func TestTokenResponseHeader(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusOK)
	}

	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatal(resp.Header.Get("Cache-Control"), "no-store")
	} else if resp.Header.Get("Pragma") != "no-cache" {
		t.Fatal(resp.Header.Get("Pragma"), "no-cache")
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
		req.Header.Set("Connection", "close")
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

		var res struct{ Error string }
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			util.LogRequest(level.ERR, req, true)
			util.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if err := json.Unmarshal(data, &res); err != nil {
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

	if res, err := testFromRequestAuthToGetAccountInfo(idpSys, cli, map[string]string{
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
	}, kid, sigKey, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != testAcc.attribute("email") {
		t.Fatal(em, testAcc.attribute("email"))
	}
}

// トークンリクエストのパラメータが重複していたら拒否できるか。
func TestDenyOverlapParameterInTokenRequest(t *testing.T) {
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

	req, err := http.NewRequest("POST", idpSys.selfId+"/token", strings.NewReader(url.Values{
		"grant_type":            {"authorization_code"},
		"redirect_uri":          {rediUri},
		"client_id":             {testTa2.id()},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"code":                  {cod},
		"client_assertion":      {ass},
	}.Encode()+"&grant_type=authorization_code"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "close")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogRequest(level.ERR, req, true)
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// トークンリクエストで grant_type が authorization_code なのに client_id が無かったら拒否できるか。
func TestDenyTokenRequestWithoutClientId(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"client_id":             "",
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// トークンリクエストに grant_type が無いなら拒否できるか。
func TestDenyNoGrantTypeInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"grant_type":            "",
		"redirect_uri":          rediUri,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// トークンリクエストに code が無いなら拒否できるか。
func TestDenyNoCodeInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"code":                  "",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// トークンリクエストに redirect_uri が無いなら拒否できるか。
func TestDenyNoRedirectUriInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"redirect_uri":          "",
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// 知らない grant_type なら拒否できるか。
func TestDenyUnknownGrantTypeInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"grant_type":            "unknown_grant_type",
		"redirect_uri":          rediUri,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errUnsuppGrntType {
		t.Fatal(res.Error, errUnsuppGrntType)
	}
}

// 認可コードが発行されたクライアントでないなら拒否できるか。
// クライアントが不正な認可コードを使っているとして error は invalid_grant か。
func TestDenyNotCodeHolder(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	taBuff := *testTa
	taBuff.Id = "http://example.org"
	testTa1 := &taBuff

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, []*ta{testTa1})
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
	}, map[string]interface{}{
		"alg": "RS256",
		"kid": kid,
	}, map[string]interface{}{
		"iss": testTa1.id(),
		"sub": testTa1.id(),
		"aud": idpSys.selfId + "/token",
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idpSys.idTokExpiDur).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          rediUri,
		"client_id":             testTa1.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvGrnt {
		t.Fatal(res.Error, res.Error_description, errInvGrnt)
	}
}

// 認可コードがおかしかったら拒否できるか。
// error は invalid_grant か。
func TestDenyInvalidCode(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"code":                  "cooooooooooooooooooooooooooooooode",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvGrnt {
		t.Fatal(res.Error, res.Error_description, errInvGrnt)
	}
}

// redirect_uri が認証リクエストのときと違っていたら拒否できるか。
// error は invalid_grant か。
func TestDenyInvalidRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/a", "/b"}, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトできたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	var rediUri2 string
	if strings.HasSuffix(rediUri, "/a") {
		rediUri2 = rediUri[:len(rediUri)-len("/a")] + "/b"
	} else {
		rediUri2 = rediUri[:len(rediUri)-len("/b")] + "/a"
	}

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"redirect_uri":          rediUri2,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvGrnt {
		t.Fatal(res.Error, res.Error_description, errInvGrnt)
	}
}

// 複数のクライアント認証方式が使われていたら違っていたら拒否できるか。
// error は invalid_request か。
func TestDenyManyClientAuthAlgorithms(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/a", "/b"}, []*account{testAcc}, nil)
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"client_secret":         "hi-mi-tsu",
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, res.Error_description, errInvReq)
	}
}

// クライアントを認証できないなら拒否できるか。
// error は invalid_client か。
func TestDenyInvalidClient(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/a", "/b"}, []*account{testAcc}, nil)
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idpSys, cli, map[string]string{
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
		"client_assertion":      "ore dayo ore ore",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvTa {
		t.Fatal(res.Error, res.Error_description, errInvTa)
	}
}

// 期限切れの認可コードを拒否できるか。
// error は invalid_grant か。
func TestDenyExpiredCode(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/a", "/b"}, []*account{testAcc}, nil)
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

	time.Sleep(idpSys.codExpiDur + time.Millisecond)

	resp, err := testGetTokenWithoutCheck(idpSys, consResp, map[string]interface{}{
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
		"client_assertion":      "ore dayo ore ore",
	}, kid, sigKey)

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvGrnt {
		t.Fatal(res.Error, res.Error_description, errInvGrnt)
	}
}

// 認可コードが 2 回使われたら拒否できるか。
func TestDenyUsedCode(t *testing.T) {
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

	// 1 回目はアクセストークンを取得できる。
	tokRes, err := testGetToken(idpSys, consResp, map[string]interface{}{
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
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	} else if tokRes["access_token"] == "" {
		t.Fatal(tokRes)
	}

	// 2 回目は拒否される。
	resp, err := testGetTokenWithoutCheck(idpSys, consResp, map[string]interface{}{
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
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvGrnt {
		t.Fatal(res.Error, errInvGrnt)
	}
}

// 2 回使われた認可コードで発行したアクセストークンを無効にできるか。
func _TestDisableTokenOfUsedCode(t *testing.T) {
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

	// アクセストークンを取得してアカウント情報も取得する。
	tokRes, err := testGetToken(idpSys, consResp, map[string]interface{}{
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
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}
	if res, err := testGetAccountInfo(idpSys, tokRes, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != testAcc.attribute("email") {
		t.Fatal(em, testAcc.attribute("email"))
	}

	// もう一度アクセストークンを要求して拒否される。
	tokResp, err := testGetTokenWithoutCheck(idpSys, consResp, map[string]interface{}{
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
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}
	tokResp.Body.Close()

	// 拒否されていることは別テスト。

	// さっき取得したアクセストークンでのアカウント情報取得も拒否される。
	resp, err := testGetAccountInfoWithoutCheck(idpSys, tokRes, nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		util.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvTok {
		t.Fatal(res.Error, errInvTok)
	}
}

// ID トークンが iss, sub, aud, exp, iat クレームを含むか。
func TestIdToken(t *testing.T) {
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

	if res, err := testFromRequestAuthToGetToken(idpSys, cli, map[string]string{
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
	}, kid, sigKey); err != nil {
		t.Fatal(err)
	} else if idTok, _ := res["id_token"].(string); idTok == "" {
		t.Fatal("no id token")
	} else if jws, err := util.ParseJws(idTok); err != nil {
		t.Fatal(err)
	} else if iss, _ := jws.Claim("iss").(string); iss != idpSys.selfId {
		t.Fatal(iss, idpSys.selfId)
	} else if sub, _ := jws.Claim("sub").(string); sub != testAcc.id() {
		t.Fatal(sub, testAcc.id())
	} else if !audienceHas(jws.Claim("aud"), testTa2.id()) {
		t.Fatal(jws.Claim("aud"), testTa2.id())
	} else if exp, _ := jws.Claim("exp").(float64); exp < float64(time.Now().Unix()) {
		t.Fatal(exp, time.Now().Unix())
	} else if iat, _ := jws.Claim("iat").(float64); iat > float64(time.Now().Unix()) {
		t.Fatal(iat, time.Now().Unix())
	}
}

// 認証リクエストが max_age パラメータを含んでいたら、ID トークンが auth_time クレームを含むか。
func TestAuthTimeOfIdToken(t *testing.T) {
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

	if res, err := testFromRequestAuthToGetToken(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"max_age":       "1",
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
	}, kid, sigKey); err != nil {
		t.Fatal(err)
	} else if idTok, _ := res["id_token"].(string); idTok == "" {
		t.Fatal("no id token")
	} else if jws, err := util.ParseJws(idTok); err != nil {
		t.Fatal(err)
	} else if at, _ := jws.Claim("auth_time").(float64); at > float64(time.Now().Unix()) {
		t.Fatal(at, time.Now().Unix())
	}
}
