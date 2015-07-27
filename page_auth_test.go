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

// 認証リクエスト・レスポンス周りのテスト。

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/rglog/level"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

// 知らないパラメータを無視できるか。
func TestIgnoreUnknownParameterInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	if res, err := testFromRequestAuthToGetAccountInfo(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"unknown_name":  "unknown_value",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss":     ta.taInfo().Id(),
		"sub":     ta.taInfo().Id(),
		"aud":     idp.sys.selfId + test_pathTok,
		"jti":     strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp":     time.Now().Add(idp.sys.jtiExpIn).Unix(),
		"unknown": "unknown",
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		"unknown":               "unknown",
	}, test_taPriKey, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != acnt.Attribute("email") {
		t.Error(em)
		t.Fatal(acnt.Attribute("email"))
	}
}

// 認証リクエストの重複パラメータを拒否できるか。
// エラーリダイレクトして error は invalid_request か。
func TestDenyOverlapParameterInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	req, err := http.NewRequest("GET", idp.sys.selfId+test_pathAuth+"?"+url.Values{
		"scope":         {"openid email"},
		"response_type": {"code"},
		"client_id":     {ta.taInfo().Id()},
		"redirect_uri":  {ta.redirectUri()},
	}.Encode()+"&scope=aaaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Connection", "close")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogRequest(level.ERR, req, true)
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if resp.Request.FormValue("error") != idperr.Invalid_request {
		t.Fatal("no error")
	}
}

// 認証リクエストに scope が無かったら拒否できるか。
// 必須パラメータ無しでエラーリダイレクトして error は invalid_request か。
func TestDenyNoScopeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuth(idp, cli, map[string]string{
		"scope":         "",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("error") != idperr.Invalid_request {
		t.Error(q.Get("error"))
		t.Fatal(idperr.Invalid_request)
	}
}

// 知らない scope 値を無視できるか。
func TestIgnoreUnknownScopes(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	if res, err := testFromRequestAuthToGetToken(idp, cli, map[string]string{
		"scope":         "openid email unknown_scope",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email unknown_scope",
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss": ta.taInfo().Id(),
		"sub": ta.taInfo().Id(),
		"aud": idp.sys.selfId + test_pathTok,
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idp.sys.jtiExpIn).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey); err != nil {
		t.Fatal(err)
	} else if scop, _ := res["scope"].(string); request.FormValueSet(scop)["unknown_scope"] {
		t.Fatal(scop)
	}
}

// 認証リクエストに client_id が無い時に拒否できるか。
func TestDenyNoClientIdInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     "",
		"redirect_uri":  ta.redirectUri(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Error(resp.Status)
		t.Fatal(http.StatusBadRequest)
	}
}

// 認証リクエストに response_type が無い時に拒否できるか。
// 必須パラメータ無しで error は invalid_request か。
func TestDenyNoResponseTypeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuth(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Request.FormValue("error") != idperr.Invalid_request {
		t.Error(resp.Request.FormValue("error"))
		t.Fatal(idperr.Invalid_request)
	}
}

// 認証リクエストの response_type が未知の時に拒否できるか。
// error は unsupported_response_type か。
func TestDenyUnknownResponseTypeInAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuth(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "unknown",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Request.FormValue("error") != idperr.Unsupported_response_type {
		t.Error(resp.Request.FormValue("error"))
		t.Fatal(idperr.Invalid_request)
	}
}

// リソースオーナーが拒否したら error は access_denied か。
func TestErrorWhenOwnerDenied(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"denied_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Request.FormValue("error") != idperr.Access_denied {
		t.Error(resp.Request.FormValue("error"))
		t.Fatal(idperr.Access_denied)
	}
}

// 結果をリダイレクトで返すときに redirect_uri のパラメータを維持できるか。
func TestKeepRedirectUriParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	ta, err := newTestTaServer("/callback?param_name=param_value")
	if err != nil {
		t.Fatal(err)
	}
	defer ta.close()
	tas := []tadb.Element{ta.taInfo()}
	acnt := newTestAccount()
	idp, err := newTestIdpServer([]account.Element{acnt}, tas, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	} else if q.Get("param_name") != "param_value" {
		t.Error(q.Get("param_name"))
		t.Fatal("param_value")
	}
}

// エラーをリダイレクトで返すときに redirect_uri のパラメータを維持できるか。
func TestKeepRedirectUriParameterInError(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	ta, err := newTestTaServer("/callback?param_name=param_value")
	if err != nil {
		t.Fatal(err)
	}
	defer ta.close()
	tas := []tadb.Element{ta.taInfo()}
	acnt := newTestAccount()
	idp, err := newTestIdpServer([]account.Element{acnt}, tas, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuth(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "unknown",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("error") != idperr.Unsupported_response_type {
		t.Error(q.Get("error"))
		t.Fatal(idperr.Unsupported_response_type)
	} else if q.Get("param_name") != "param_value" {
		t.Error(q.Get("param_name"))
		t.Fatal("param_value")
	}
}

// redirect_uri が登録値と異なるときにリダイレクトせずに拒否できるか。
func TestDirectErrorResponseInInvalidRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri() + "/a",
	})
	defer resp.Body.Close()

	// エラー UI にリダイレクトされる。
	if resp.StatusCode != http.StatusBadRequest {
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}
}

// redirect_uri が無いときにリダイレクトせずに拒否できるか。
func TestDirectErrorResponseInNoRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Error(resp.Status)
		t.Fatal(http.StatusBadRequest)
	}
}

// 結果をリダイレクトで返すときに state パラメータも返せるか。
func TestReturnStateParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"state":         "test_state",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
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
		t.Error(q.Get("state"))
		t.Fatal("test_state")
	}
}

// エラーをリダイレクトで返すときに state パラメータも返せるか。
func TestReturnStateAtError(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testRequestAuth(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "unknown",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"state":         "test_state",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if q := resp.Request.URL.Query(); q.Get("error") != idperr.Unsupported_response_type {
		t.Error(q.Get("error"))
		t.Fatal(idperr.Unsupported_response_type)
	} else if q.Get("state") != "test_state" {
		t.Error(q.Get("state"))
		t.Fatal("test_state")
	}
}

// POST での認証リクエストにも対応するか。
func TestPostAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	q := url.Values{}
	q.Set("scope", "openid email")
	q.Set("response_type", "code")
	q.Set("client_id", ta.taInfo().Id())
	q.Set("redirect_uri", ta.redirectUri())
	q.Set("prompt", "select_account login consent")
	req, err := http.NewRequest("POST", idp.sys.selfId+"/auth", strings.NewReader(q.Encode()))
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
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	}
}

// prompt が login を含むなら認証させるか。
func TestForceLogin(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 一旦認証を通す。
	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "login",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	consResp.Body.Close()
	if q := consResp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "login",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if resp.Request.URL.Path != test_pathLginUi {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.Request.URL.Path)
		t.Fatal(test_pathLginUi)
	}
}

// prompt が none と login を含むならエラーを返すか。
func TestForceLoginError(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "none login",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if q := resp.Request.URL.Query(); q.Get("error") != idperr.Login_required {
		t.Error(q.Get("error"))
		t.Fatal(idperr.Login_required)
	}
}

// prompt が consent を含むなら同意させるか。
func TestForceConsent(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 一旦認証を通す。
	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "consent",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer consResp.Body.Close()
	if q := consResp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	}

	// 同意 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "consent",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if resp.Request.URL.Path != test_pathConsUi {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.Request.URL.Path)
		t.Fatal(test_pathConsUi)
	}
}

// prompt が none と consent を含むならエラーを返すか。
func TestForceConsentError(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 一旦認証を通す。
	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "consent",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer consResp.Body.Close()
	if q := consResp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "none consent",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if q := resp.Request.URL.Query(); q.Get("error") != idperr.Consent_required {
		t.Error(q.Get("error"))
		t.Fatal(idperr.Consent_required)
	}
}

// prompt が select_account を含むならアカウント選択させるか。
func TestForceSelect(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// アカウント選択 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "select_account",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if resp.Request.URL.Path != test_pathSelUi {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.Request.URL.Path)
		t.Fatal(test_pathSelUi)
	}
}

// prompt が none と select_account を含むならエラーを返すか。
func TestForceSelectError(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "none select_account",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if q := resp.Request.URL.Query(); q.Get("error") != idperr.Account_selection_required {
		t.Error(q.Get("error"))
		t.Fatal(idperr.Account_selection_required)
	}
}

// 最後に認証してから max_age パラメータの値より時間が経っているときに認証させるか。
func TestLoginTimeout(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 一旦認証を通す。
	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "login",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	})
	if err != nil {
		t.Fatal(err)
	}
	consResp.Body.Close()
	if q := consResp.Request.URL.Query(); q.Get("code") == "" {
		t.Fatal("no code")
	}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"max_age":       "0",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if resp.Request.URL.Path != test_pathLginUi {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.Request.URL.Path)
		t.Fatal(test_pathLginUi)
	}
}

// UI にパラメータが渡せてるか。
func TestUiParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	// 認証 UI に飛ばされる。
	resp, err := testRequestAuthWithoutCheck(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "login",
		"display":       "page",
		"ui_locales":    "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if resp.Request.URL.Path != test_pathLginUi {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.Request.URL.Path)
		t.Fatal(test_pathLginUi)
	} else if q := resp.Request.URL.Query(); q.Get("display") != "page" {
		t.Error(q.Get("display"))
		t.Fatal("page")
	} else if q.Get("locales") != "ja" {
		t.Error(q.Get("locales"))
		t.Fatal("ja")
	}
}

// claims パラメータを処理できるか。
func TestClaimsParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	clms, err := json.Marshal(map[string]interface{}{
		"userinfo": map[string]interface{}{
			"email": nil,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res, err := testFromRequestAuthToGetAccountInfo(idp, cli, map[string]string{
		"scope":         "openid",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"claims":        string(clms),
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope":  "openid",
		"allowed_claims": "email",
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss": ta.taInfo().Id(),
		"sub": ta.taInfo().Id(),
		"aud": idp.sys.selfId + test_pathTok,
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idp.sys.jtiExpIn).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != acnt.Attribute("email") {
		t.Error(em)
		t.Fatal(acnt.Attribute("email"))
	}
}

// essential クレームを拒否されたら拒否できるか。
func TestDenyDeniedEssentialClaim(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"claims":        `{"userinfo":{"email":{"essential":true}}}`,
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid",
		"denied_claim":  "email",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if q := resp.Request.URL.Query(); q.Get("error") != idperr.Access_denied {
		server.LogRequest(level.ERR, resp.Request, true)
		t.Error(q.Get("error"))
		t.Fatal(idperr.Access_denied)
	}
}

// 値の違う sub クレームを要求されたら拒否できるか。
func TestDenyInvalidSubClaim(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	resp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
		"scope":         "openid",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"claims":        `{"userinfo":{"sub":{"value":"` + acnt.Id() + `a"}}}`,
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	} else if q := resp.Request.URL.Query(); q.Get("error") != idperr.Access_denied {
		server.LogRequest(level.ERR, resp.Request, true)
		t.Error(q.Get("error"))
		t.Fatal(idperr.Access_denied)
	}
}

// request パラメータを受け取れるか。
func TestRequestParam(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
	defer ta.close()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	jt := jwt.New()
	jt.SetHeader("alg", "none")
	jt.SetClaim("scope", "openid")
	jt.SetClaim("response_type", "code")
	jt.SetClaim("redirect_uri", ta.redirectUri())
	jt.SetClaim("claims", map[string]interface{}{
		"userinfo": map[string]interface{}{
			"email": map[string]interface{}{
				"essential": true,
			},
		},
	})
	buff, err := jt.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if res, err := testFromRequestAuthToGetAccountInfo(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"request":       string(buff),
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "password",
		"password":  test_acntPasswd,
	}, map[string]string{
		"allowed_scope": "openid email",
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss": ta.taInfo().Id(),
		"sub": ta.taInfo().Id(),
		"aud": idp.sys.selfId + test_pathTok,
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idp.sys.jtiExpIn).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != acnt.Attribute("email") {
		t.Error(em)
		t.Fatal(acnt.Attribute("email"))
	}
}
