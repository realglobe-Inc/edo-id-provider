package main

// 認証リクエスト・レスポンス周りのテスト。

import (
	"encoding/json"
	logutil "github.com/realglobe-Inc/edo/util/log"
	"github.com/realglobe-Inc/edo/util/server"
	"github.com/realglobe-Inc/edo/util/strset"
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
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

// 知らないパラメータを無視できるか。
func TestIgnoreUnknownParameterInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
// エラーリダイレクトして error は invalid_request か。
func TestDenyOverlapParameterInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
	req.Header.Set("Connection", "close")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogRequest(level.ERR, req, true)
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusOK)
	} else if resp.Request.FormValue(formErr) != errInvReq {
		t.Fatal("no error")
	}
}

// 認証リクエストに scope が無かったら拒否できるか。
// 必須パラメータ無しでエラーリダイレクトして error は invalid_request か。
func TestDenyNoScopeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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

	if q := resp.Request.URL.Query(); q.Get("error") != errInvReq {
		t.Fatal(q.Get("error"), errInvReq)
	}
}

// 知らない scope 値を無視できるか。
func TestIgnoreUnknownScopes(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	if res, err := testFromRequestAuthToGetToken(idpSys, cli, map[string]string{
		"scope":         "openid email unknown_scope",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	}, map[string]string{
		"username": testAcc.name(),
	}, map[string]string{
		"username": testAcc.name(),
		"password": testAcc.password(),
	}, map[string]string{
		"consented_scope": "openid email unknown_scope",
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
	} else if scop, _ := res["scope"].(string); strset.FromSlice(strings.Split(scop, " "))["unknown_scope"] {
		t.Fatal(scop)
	}
}

// 認証リクエストに client_id が無い時に拒否できるか。
// 必須パラメータ無しで error は invalid_request か。
func TestDenyNoClientIdInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	_, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// 認証リクエストに response_type が無い時に拒否できるか。
// 必須パラメータ無しで error は invalid_request か。
func TestDenyNoResponseTypeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
// error は unsupported_response_type か。
func TestDenyUnknownResponseTypeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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

// リソースオーナーが拒否したら error は access_denied か。
func TestErrorWhenOwnerDenied(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/redirect_endpoint?param_name=param_value"}, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
		"denied_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Request.FormValue(formErr) != errAccDeny {
		t.Fatal(resp.Request.FormValue(formErr), errAccDeny)
	}
}

// 結果をリダイレクトで返すときに redirect_uri のパラメータを維持できるか。
func TestKeepRedirectUriParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/redirect_endpoint?param_name=param_value"}, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp([]string{"/redirect_endpoint?param_name=param_value"}, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// redirect_uri が無いときにリダイレクトせずに拒否できるか。
func TestDirectErrorResponseInNoRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, _, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != errInvReq {
		t.Fatal(res.Error, errInvReq)
	}
}

// 結果をリダイレクトで返すときに state パラメータも返せるか。
func TestReturnStateParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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

// エラーをリダイレクトで返すときに state パラメータも返せるか。
func TestReturnStateAtError(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
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
		"state":         "test_state",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("error") != errUnsuppRespType {
		t.Fatal(q.Get("error"), errUnsuppRespType)
	} else if q.Get("state") != "test_state" {
		t.Fatal(q.Get("state"), "test_state")
	}
}

// POST での認証リクエストにも対応するか。
func TestPostAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	q := url.Values{}
	q.Set("scope", "openid email")
	q.Set("response_type", "code")
	q.Set("client_id", testTa2.id())
	q.Set("redirect_uri", rediUri)
	q.Set("prompt", "select_account login consent")
	req, err := http.NewRequest("POST", idpSys.selfId+"/auth", strings.NewReader(q.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "close")
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusOK)
	}
}

// prompt が login を含むなら認証させるか。
func TestForceLogin(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, _, _, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	// サーバ起動待ち。
	time.Sleep(10 * time.Millisecond)

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	// 一旦認証を通す。
	consResp, err := testFromRequestAuthToConsent(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"prompt":        "login",
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
	consResp.Body.Close()
	if q := consResp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
		"prompt":        "login",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.StatusCode, http.StatusOK)
	} else if resp.Request.URL.Path != idpSys.uiUri+"/login.html" {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(resp.Request.URL.Path, idpSys.uiUri+"/login.html")
	}
}
