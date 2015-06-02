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

package token

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

func newTestHandler(keys []jwk.Key, acnts []account.Element, tas []tadb.Element) *handler {
	return New(
		server.NewStopper(),
		"https://idp.example.org",
		"ES256",
		"",
		"/token",
		20,
		30,
		time.Minute,
		time.Hour,
		time.Minute,
		keydb.NewMemoryDb(keys),
		account.NewMemoryDb(acnts),
		tadb.NewMemoryDb(tas),
		sector.NewMemoryDb(),
		pairwise.NewMemoryDb(),
		authcode.NewMemoryDb(),
		token.NewMemoryDb(),
		jtidb.NewMemoryDb(),
		rand.New(time.Second),
	).(*handler)
}

// 正常系。
// レスポンスが access_token, token_type, expires_in を含むことの検査。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
func TestNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	} else if contType, contType2 := "application/json", w.HeaderMap.Get("Content-Type"); contType2 != contType {
		t.Error(contType2)
		t.Fatal(contType)
	} else if cc, cc2 := "no-store", w.HeaderMap.Get("Cache-Control"); cc2 != cc {
		t.Error(cc2)
		t.Fatal(cc)
	} else if prgm, prgm2 := "no-cache", w.HeaderMap.Get("Pragma"); prgm2 != prgm {
		t.Error(prgm2)
		t.Fatal(prgm)
	}

	var buff struct {
		Access_token string
		Token_type   string
		Expires_in   int
		Id_token     string
	}
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Access_token == "" {
		t.Fatal("no access token")
	} else if tokType := "Bearer"; buff.Token_type != tokType {
		t.Error(buff.Token_type)
		t.Fatal(tokType)
	} else if expIn := int(hndl.tokExpIn / time.Second); buff.Expires_in != expIn {
		t.Error(buff.Expires_in)
		t.Fatal(expIn)
	}
	idTok, err := jwt.Parse([]byte(buff.Id_token))
	if err != nil {
		t.Fatal(err)
	} else if alg, _ := idTok.Header("alg").(string); alg != hndl.sigAlg {
		t.Error(alg)
		t.Fatal(hndl.sigAlg)
	}
}

// POST でないリクエストを拒否できることの検査。
func TestDenyNonPost(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		r.Method = "GET"
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		r.URL.RawQuery = string(buff)
		r.Body = nil
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// リクエストの未知のパラメータを無視できることの検査。
func TestIgnoreUnknownParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Set("unknown", "unknown")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	}
	var buff struct{ Access_token string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Access_token == "" {
		t.Fatal("no access token")
	}
}

// リクエストのパラメータが重複していたら拒否できることの検査。
func TestDenyOverlappedParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode() + "&grant_type=authorization_code"))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// リクエストで grant_type が authorization_code なのに
// client_id が無かったら拒否できることの検査。
func TestDenyNoClientId(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Del("client_id")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// リクエストに grant_type が無いなら拒否できることの検査。
func TestDenyNoGrantType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Del("grant_type")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// リクエストに code が無いなら拒否できることの検査。
func TestDenyNoCode(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Del("code")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// リクエストに redirect_uri が無いなら拒否できることの検査。
func TestDenyNoRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Del("redirect_uri")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 知らない grant_type を拒否できることの検査。
func TestDenyUnknownGrantType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Set("grant_type", "unknown")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "unsupported_grant_type"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 認可コードが発行されたクライアントでないなら拒否できることの検査。
func TestDenyNonCodeHolder(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	ta := tadb.New(test_ta.Id()+"a", nil, strsetutil.New(test_rediUri), []jwk.Key{test_taKey}, false, "")

	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta, ta})

	cod := newCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Set("client_id", ta.Id())
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusBadRequest)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "invalid_grant"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}
