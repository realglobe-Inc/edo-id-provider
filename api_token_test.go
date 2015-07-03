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

// トークンリクエスト・レスポンス周りのテスト。

import (
	"bytes"
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/hash"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

// トークンレスポンスが access_token, token_type, expires_in を含むか。
func TestTokenResponse(t *testing.T) {
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
		"allowed_scope": "openid",
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
	} else if scop, _ := res["scope"].(string); scop == "" {
		t.Fatal("no scope")
	}
}

// トークンレスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むか。
func TestTokenResponseHeader(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusOK)
	}

	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Error(resp.Header.Get("Cache-Control"))
		t.Fatal("no-store")
	} else if resp.Header.Get("Pragma") != "no-cache" {
		t.Error(resp.Header.Get("Pragma"))
		t.Fatal("no-cache")
	}
}

// POST でないトークンリクエストを拒否できるか。
func TestDenyNonPostTokenRequest(t *testing.T) {
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

	for _, meth := range []string{"GET", "PUT"} {
		// TA にリダイレクトしたときのレスポンスを設定しておく。
		ta.AddResponse(http.StatusOK, nil, []byte("success"))

		cook, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		cli := &http.Client{Jar: cook}

		consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
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
		defer consResp.Body.Close()

		cod := consResp.Request.FormValue("code")
		if cod == "" {
			server.LogRequest(level.ERR, consResp.Request, true)
			t.Fatal("no code")
		}

		// 認可コードを取得できた。

		assJt := jwt.New()
		assJt.SetHeader("alg", "ES384")
		assJt.SetClaim("iss", ta.taInfo().Id())
		assJt.SetClaim("sub", ta.taInfo().Id())
		assJt.SetClaim("aud", idp.sys.selfId+test_pathTok)
		assJt.SetClaim("jti", strconv.FormatInt(time.Now().UnixNano(), 16))
		assJt.SetClaim("exp", time.Now().Add(idp.sys.jtiExpIn).Unix())
		assJt.SetClaim("code", cod)
		if err := assJt.Sign([]jwk.Key{test_taPriKey}); err != nil {
			t.Fatal(err)
		}
		assBuff, err := assJt.Encode()
		if err != nil {
			t.Fatal(err)
		}
		ass := string(assBuff)

		req, err := http.NewRequest(meth, idp.sys.selfId+test_pathTok, strings.NewReader(url.Values{
			"grant_type":            {"authorization_code"},
			"redirect_uri":          {ta.redirectUri()},
			"client_id":             {ta.taInfo().Id()},
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
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Error(resp.StatusCode)
			t.Fatal(http.StatusMethodNotAllowed)
		}

		var res struct{ Error string }
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if err := json.Unmarshal(data, &res); err != nil {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if res.Error != idperr.Invalid_request {
			t.Error(res.Error)
			t.Fatal(idperr.Invalid_request)
		}
	}
}

// トークンリクエストの未知のパラメータを無視できるか。
func TestIgnoreUnknownParameterInTokenRequest(t *testing.T) {
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
		"unknown_name":          "unknown_value",
	}, test_taPriKey, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != acnt.Attribute("email") {
		t.Error(em)
		t.Fatal(acnt.Attribute("email"))
	}
}

// トークンリクエストのパラメータが重複していたら拒否できるか。
func TestDenyOverlapParameterInTokenRequest(t *testing.T) {
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

	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
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
	defer consResp.Body.Close()

	cod := consResp.Request.FormValue("code")
	if cod == "" {
		server.LogRequest(level.ERR, consResp.Request, true)
		t.Fatal("no code")
	}

	// 認可コードを取得できた。

	assJt := jwt.New()
	assJt.SetHeader("alg", "ES384")
	assJt.SetClaim("iss", ta.taInfo().Id())
	assJt.SetClaim("sub", ta.taInfo().Id())
	assJt.SetClaim("aud", idp.sys.selfId+test_pathTok)
	assJt.SetClaim("jti", strconv.FormatInt(time.Now().UnixNano(), 16))
	assJt.SetClaim("exp", time.Now().Add(idp.sys.jtiExpIn).Unix())
	assJt.SetClaim("code", cod)
	if err := assJt.Sign([]jwk.Key{test_taPriKey}); err != nil {
		t.Fatal(err)
	}
	assBuff, err := assJt.Encode()
	if err != nil {
		t.Fatal(err)
	}
	ass := string(assBuff)

	req, err := http.NewRequest("POST", idp.sys.selfId+test_pathTok, strings.NewReader(url.Values{
		"grant_type":            {"authorization_code"},
		"redirect_uri":          {ta.redirectUri()},
		"client_id":             {ta.taInfo().Id()},
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
		server.LogRequest(level.ERR, req, true)
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogRequest(level.ERR, req, true)
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogRequest(level.ERR, req, true)
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_request {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_request)
	}
}

// トークンリクエストで grant_type が authorization_code なのに client_id が無かったら拒否できるか。
func TestDenyTokenRequestWithoutClientId(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"client_id":             "",
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_request {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_request)
	}
}

// トークンリクエストに grant_type が無いなら拒否できるか。
func TestDenyNoGrantTypeInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss": ta.taInfo().Id(),
		"sub": ta.taInfo().Id(),
		"aud": idp.sys.selfId + test_pathTok,
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idp.sys.jtiExpIn).Unix(),
	}, map[string]string{
		"grant_type":            "",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_request {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_request)
	}
}

// トークンリクエストに code が無いなら拒否できるか。
func TestDenyNoCodeInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"code":                  "",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_request {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_request)
	}
}

// トークンリクエストに redirect_uri が無いなら拒否できるか。
func TestDenyNoRedirectUriInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"redirect_uri":          "",
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_request {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_request)
	}
}

// 知らない grant_type なら拒否できるか。
func TestDenyUnknownGrantTypeInTokenRequest(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss": ta.taInfo().Id(),
		"sub": ta.taInfo().Id(),
		"aud": idp.sys.selfId + test_pathTok,
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idp.sys.jtiExpIn).Unix(),
	}, map[string]string{
		"grant_type":            "unknown_grant_type",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Unsupported_grant_type {
		t.Error(res.Error)
		t.Fatal(idperr.Unsupported_grant_type)
	}
}

// 認可コードが発行されたクライアントでないなら拒否できるか。
// クライアントが不正な認可コードを使っているとして error は invalid_grant か。
func TestDenyNotCodeHolder(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	ta2, err := newTestTaServer("")
	if err != nil {
		t.Fatal(err)
	}
	defer ta2.close()
	tas := []tadb.Element{ta2.taInfo()}

	acnt := newTestAccount()
	idp, ta, err := setupTestIdpAndTa([]account.Element{acnt}, tas, nil, nil)
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
	}, map[string]interface{}{
		"alg": "ES384",
	}, map[string]interface{}{
		"iss": ta2.taInfo().Id(),
		"sub": ta2.taInfo().Id(),
		"aud": idp.sys.selfId + test_pathTok,
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idp.sys.jtiExpIn).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          ta.redirectUri(),
		"client_id":             ta2.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_grant {
		t.Error(res.Error, res.Error_description)
		t.Fatal(idperr.Invalid_grant)
	}
}

// 認可コードがおかしかったら拒否できるか。
// error は invalid_grant か。
func TestDenyInvalidCode(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"code":                  "cooooooooooooooooooooooooooooooode",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_grant {
		t.Error(res.Error, res.Error_description)
		t.Fatal(idperr.Invalid_grant)
	}
}

// redirect_uri が認証リクエストのときと違っていたら拒否できるか。
// error は invalid_grant か。
func TestDenyInvalidRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	ta, err := newTestTaServer(test_taPathCb + "2")
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"redirect_uri":          ta.taInfo().Id() + test_taPathCb,
		"client_id":             ta.taInfo().Id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_grant {
		t.Error(res.Error, res.Error_description)
		t.Fatal(idperr.Invalid_grant)
	}
}

// 複数のクライアント認証方式が使われていたら拒否できるか。
// error は invalid_request か。
func TestDenyManyClientAuthAlgorithms(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"client_secret":         "hi-mi-tsu",
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_request {
		t.Error(res.Error, res.Error_description)
		t.Fatal(idperr.Invalid_request)
	}
}

// クライアントを認証できないなら拒否できるか。
// error は invalid_client か。
func TestDenyInvalidClient(t *testing.T) {
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

	resp, err := testFromRequestAuthToGetTokenWithoutCheck(idp, cli, map[string]string{
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
		"client_assertion":      "ore dayo ore ore",
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_client {
		t.Error(res.Error, res.Error_description)
		t.Fatal(idperr.Invalid_client)
	}
}

// 期限切れの認可コードを拒否できるか。
// error は invalid_grant か。
func TestDenyExpiredCode(t *testing.T) {
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

	codExpIn := time.Millisecond
	idp.authPage.SetCodeExpiresIn(codExpIn)
	// 同期。
	idp.sys.stopper.Lock()
	idp.sys.stopper.Unlock()

	// TA にリダイレクトしたときのレスポンスを設定しておく。
	ta.AddResponse(http.StatusOK, nil, []byte("success"))

	cook, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cook}

	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
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

	time.Sleep(codExpIn + time.Millisecond)

	resp, err := testGetTokenWithoutCheck(idp, consResp, map[string]interface{}{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error, Error_description string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_grant {
		t.Error(res.Error, res.Error_description)
		t.Fatal(idperr.Invalid_grant)
	}
}

// 認可コードが 2 回使われたら拒否できるか。
func TestDenyUsedCode(t *testing.T) {
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

	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
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
	defer consResp.Body.Close()

	// 1 回目はアクセストークンを取得できる。
	tokRes, err := testGetToken(idp, consResp, map[string]interface{}{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	} else if tokRes["access_token"] == "" {
		t.Fatal(tokRes)
	}

	// 2 回目は拒否される。
	resp, err := testGetTokenWithoutCheck(idp, consResp, map[string]interface{}{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_grant {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_grant)
	}
}

// 2 回使われた認可コードで発行したアクセストークンを無効にできるか。
func TestDisableTokenOfUsedCode(t *testing.T) {
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

	consResp, err := testFromRequestAuthToConsent(idp, cli, map[string]string{
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
	defer consResp.Body.Close()

	// アクセストークンを取得してアカウント情報も取得する。
	tokRes, err := testGetToken(idp, consResp, map[string]interface{}{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}
	if res, err := testGetAccountInfo(idp, tokRes, nil); err != nil {
		t.Fatal(err)
	} else if em, _ := res["email"].(string); em != acnt.Attribute("email") {
		t.Error(em)
		t.Fatal(acnt.Attribute("email"))
	}

	// もう一度アクセストークンを要求して拒否される。
	tokResp, err := testGetTokenWithoutCheck(idp, consResp, map[string]interface{}{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	}
	tokResp.Body.Close()

	// 拒否されていることは別テスト。

	// さっき取得したアクセストークンでのアカウント情報取得も拒否される。
	resp, err := testGetAccountInfoWithoutCheck(idp, tokRes, nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		server.LogResponse(level.ERR, resp, true)
		t.Error(resp.StatusCode)
		t.Fatal(http.StatusBadRequest)
	}

	var res struct{ Error string }
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if err := json.Unmarshal(data, &res); err != nil {
		server.LogResponse(level.ERR, resp, true)
		t.Fatal(err)
	} else if res.Error != idperr.Invalid_token {
		t.Error(res.Error)
		t.Fatal(idperr.Invalid_token)
	}
}

// ID トークンが iss, sub, aud, exp, iat クレームを含むか。
func TestIdToken(t *testing.T) {
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
	} else if idTok, _ := res["id_token"].(string); idTok == "" {
		t.Fatal("no id token")
	} else if jt, err := jwt.Parse([]byte(idTok)); err != nil {
		t.Fatal(err)
	} else if err := jt.Verify([]jwk.Key{test_idpPubKey}); err != nil {
		t.Fatal(err)
	} else if iss, _ := jt.Claim("iss").(string); iss != idp.sys.selfId {
		t.Error(iss)
		t.Fatal(idp.sys.selfId)
	} else if sub, _ := jt.Claim("sub").(string); sub == "" {
		t.Fatal("no accout id")
	} else if pw, err := idp.sys.pwDb.GetByPairwise(ta.taInfo().Sector(), sub); err != nil {
		t.Fatal(err)
	} else if pw.Account() != acnt.Id() {
		t.Error(pw.Account())
		t.Fatal(acnt.Id())
	} else if !audienceHas(jt.Claim("aud"), ta.taInfo().Id()) {
		t.Error(jt.Claim("aud"))
		t.Fatal(ta.taInfo().Id())
	} else if exp, _ := jt.Claim("exp").(float64); exp < float64(time.Now().Unix()) {
		t.Error(exp)
		t.Fatal(time.Now().Unix())
	} else if iat, _ := jt.Claim("iat").(float64); iat > float64(time.Now().Unix()) {
		t.Error(iat)
		t.Fatal(time.Now().Unix())
	}
}

// 認証リクエストが max_age パラメータを含んでいたら、ID トークンが auth_time クレームを含むか。
func TestAuthTimeOfIdToken(t *testing.T) {
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
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"max_age":       "1",
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
	}, test_taPriKey); err != nil {
		t.Fatal(err)
	} else if idTok, _ := res["id_token"].(string); idTok == "" {
		t.Fatal("no id token")
	} else if jt, err := jwt.Parse([]byte(idTok)); err != nil {
		t.Fatal(err)
	} else if err := jt.Verify([]jwk.Key{test_idpPubKey}); err != nil {
		t.Fatal(err)
	} else if at, _ := jt.Claim("auth_time").(float64); at > float64(time.Now().Unix()) {
		t.Error(at)
		t.Fatal(time.Now().Unix())
	}
}

// 認証リクエストが nonce パラメータを含んでいたら、ID トークンがその値を nonce クレームとして含むか。
func TestNonceOfIdToken(t *testing.T) {
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
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"nonce":         "nonce nansu",
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
	}, test_taPriKey); err != nil {
		t.Fatal(err)
	} else if idTok, _ := res["id_token"].(string); idTok == "" {
		t.Fatal("no id token")
	} else if jt, err := jwt.Parse([]byte(idTok)); err != nil {
		t.Fatal(err)
	} else if err := jt.Verify([]jwk.Key{test_idpPubKey}); err != nil {
		t.Fatal(err)
	} else if nonc, _ := jt.Claim("nonce").(string); nonc != "nonce nansu" {
		t.Error(nonc)
		t.Fatal("nonce nansu")
	}
}

// ID トークンが署名されているか。
// ついでに at_hash クレームを含むか。
func TestIdTokenSign(t *testing.T) {
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

	res, err := testFromRequestAuthToGetToken(idp, cli, map[string]string{
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
	}, test_taPriKey)
	if err != nil {
		t.Fatal(err)
	} else if tok, _ := res["access_token"].(string); tok == "" {
		t.Fatal("no access token")
	} else if idTok, _ := res["id_token"].(string); idTok == "" {
		t.Fatal("no id token")
	}
	jt, err := jwt.Parse([]byte(res["id_token"].(string)))
	if err != nil {
		t.Fatal(err)
	} else if err := jt.Verify([]jwk.Key{test_idpPubKey}); err != nil {
		t.Fatal(err)
	} else if alg, _ := jt.Header("alg").(string); alg == "" || alg == "none" {
		t.Fatal("none sign algorithm " + alg)
	}
	hGen := jwt.HashGenerator(jt.Header("alg").(string))
	if !hGen.Available() {
		t.Error(hGen)
		t.Fatal("unsupported algorithm ", jt.Header("alg"))
	}
	hVal := hash.Hashing(hGen.New(), []byte(res["access_token"].(string)))
	hVal = hVal[:len(hVal)/2]
	ah, _ := jt.Claim("at_hash").(string)
	if ah == "" {
		t.Fatal("no at_hash")
	} else if buff, err := base64url.DecodeString(ah); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(buff, hVal) {
		t.Error(buff)
		t.Fatal(hVal)
	}
}
