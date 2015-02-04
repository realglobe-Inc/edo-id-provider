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
