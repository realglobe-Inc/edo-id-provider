package main

// 認証リクエスト・レスポンス周りのテスト。

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
)

func init() {
	util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
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

	if res, err := testFromRequestAuthToGetAccountInfo(idpSys, cli, map[string]string{
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
	}, kid, sigKey, nil); err != nil {
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

	req, err := http.NewRequest("GET", idpSys.selfId+"/auth?"+url.Values{
		"scope":         {"openid email"},
		"response_type": {"code"},
		"client_id":     {testTa2.id()},
		"redirect_uri":  {rediUri},
	}.Encode()+"&scope=aaaa", nil)
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

// 認証リクエストに scope が無かったら拒否できるか。
func TestDenyNoScopeInAuthRequest(t *testing.T) {
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
		"scope":         "",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("error") == errInvReq {
		t.Fatal(q.Get("error"), errInvReq)
	}
}

// 認証リクエストに client_id が無い時に拒否できるか。
func TestDenyNoClientIdInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	_, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
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

	resp, err := testRequestAuthWithoutCheck(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     "",
		"redirect_uri":  rediUri,
	})
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

	resp, err := testRequestAuthWithoutCheck(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri + "/a",
	})
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

	resp, err := testRequestAuthWithoutCheck(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
	})
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

// 結果をリダイレクトで返すときに state パラメータも返せるか。
func TestReturnStateParameter(t *testing.T) {
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

	resp, err := testFromRequestAuthToConsent(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"state":         "test_state",
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

	if q := resp.Request.URL.Query(); q.Get("error") != "" {
		t.Fatal(q.Get("error"))
	} else if q.Get("state") == "" {
		t.Fatal("no state")
	} else if q.Get("state") != "test_state" {
		t.Fatal(q.Get("state"), "test_state")
	}
}
