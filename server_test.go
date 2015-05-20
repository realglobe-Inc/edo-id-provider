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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

// 起動しただけでパニックを起こさないこと。
func TestBoot(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	idp, err := newTestIdpServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer idp.close()
}

// 認証してアカウント情報を取得できるか。
func TestSuccess(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
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
		"prompt":        "select_account login consent",
	}, map[string]string{
		"username": acnt.Name(),
	}, map[string]string{
		"username":  acnt.Name(),
		"pass_type": "STR43",
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
		t.Fatal(em, acnt.Attribute("email"))
	}
}

// 認証中にエラーが起きたら認証経過を破棄できるか。
func TestAbortSession(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
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

	// リクエストする。
	authResp, err := testRequestAuth(idp, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     ta.taInfo().Id(),
		"redirect_uri":  ta.redirectUri(),
		"prompt":        "select_account",
		"unknown":       "unknown",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer authResp.Body.Close()

	// アカウント選択でチケットを渡さないで認証経過をリセット。
	selResp, err := testSelectAccount(idp, cli, authResp, map[string]string{
		"username": acnt.Name(),
		"ticket":   "",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer selResp.Body.Close()

	if selResp.Request.FormValue(tagError) != idperr.Access_denied {
		t.Error(selResp.Request.FormValue(tagError))
		t.Fatal(idperr.Access_denied)
	}

	// アカウント選択でさっきのチケットを渡す。
	resp, err := testSelectAccountWithoutCheck(idp, cli, authResp, map[string]string{
		"username": acnt.Name(),
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
