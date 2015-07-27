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

// アカウント情報リクエスト・レスポンス周りのテスト。

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/rglog/level"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

// GET と POST でのアカウント情報リクエストに対応するか。
// Bearer 認証に対応するか。
// JSON を application/json で返すか。
// sub クレームを含むか。
func TestAccountInfo(t *testing.T) {
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

	tokRes, err := testFromRequestAuthToGetToken(idp, cli, map[string]string{
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

	tok, _ := tokRes["access_token"].(string)
	if tok == "" {
		t.Fatal("no token")
	}

	for _, meth := range []string{"GET", "POST"} {
		req, err := http.NewRequest(meth, idp.sys.selfId+test_pathAcnt, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
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
		} else if resp.Header.Get("Content-Type") != "application/json" {
			t.Error(resp.Header.Get("Content-Type"))
			t.Fatal("application/json")
		}

		var res struct{ Sub, Email string }
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if err := json.Unmarshal(data, &res); err != nil {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if res.Sub == "" {
			t.Fatal(res.Sub)
		} else if em, _ := acnt.Attribute("email").(string); res.Email != em {
			t.Error(res.Email)
			t.Fatal(em)
		}
	}
}
