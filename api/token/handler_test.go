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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/hash"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/rglog/level"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
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
		rand.New(time.Minute),
		true,
	).(*handler)
}

// 正常系。
// レスポンスが access_token, token_type, expires_in を含むことの検査。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
// レスポンスが scope を含むことの検査。
// ID トークンが署名されていることの検査。
// ID トークンが iss, sub, aud, exp, iat クレームを含むことの検査。
// ID トークンが nonce クレームを含むことの検査。
// ID トークンが auth_time クレームを含むことの検査。
// ID トークンが at_hash クレームを含むことの検査。
func TestNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
		Scope        string
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
	} else if scop := requtil.FormValueSet(buff.Scope); !reflect.DeepEqual(scop, cod.Scope()) {
		t.Error(scop)
		t.Fatal(cod.Scope())
	}
	idTok, err := jwt.Parse([]byte(buff.Id_token))
	if err != nil {
		t.Fatal(err)
	} else if alg, _ := idTok.Header("alg").(string); alg != hndl.sigAlg {
		t.Error(alg)
		t.Fatal(hndl.sigAlg)
	} else if !idTok.IsSigned() {
		t.Fatal("not signed ID token")
	} else if err := idTok.Verify([]jwk.Key{test_idpKey}); err != nil {
		t.Fatal(err)
	} else if iss, _ := idTok.Claim("iss").(string); iss != hndl.selfId {
		t.Error(iss)
		t.Fatal(hndl.selfId)
	} else if sub, _ := idTok.Claim("sub").(string); sub != acnt.Id() {
		t.Error(sub)
		t.Fatal(acnt.Id())
	} else if aud, ok := idTok.Claim("aud").(string); ok && aud != test_ta.Id() {
		t.Error(aud)
		t.Fatal(test_ta.Id())
	} else if aud, ok := idTok.Claim("aud").([]interface{}); ok && aud[0] != test_ta.Id() {
		t.Error(aud[0])
		t.Fatal(test_ta.Id())
	} else if exp, _ := idTok.Claim("exp").(float64); exp == 0 {
		t.Fatal("no exp")
	} else if iat, _ := idTok.Claim("iat").(float64); iat == 0 {
		t.Fatal("no iat")
	} else if exp < iat {
		t.Error("exp before iat")
		t.Error(exp)
		t.Fatal(iat)
	} else if nonc, _ := idTok.Claim("nonce").(string); nonc != test_nonc {
		t.Error(nonc)
		t.Fatal(test_nonc)
	} else if lginDate, _ := idTok.Claim("auth_time").(float64); lginDate == 0 {
		t.Fatal("no auth_time")
	} else if lginDate > iat {
		t.Error("auth_time after iat")
		t.Error(lginDate)
		t.Fatal(iat)
	}
	atHash := hash.Hashing(jwt.HashGenerator(hndl.sigAlg).New(), []byte(buff.Access_token))
	atHash = atHash[:len(atHash)/2]
	if rawAtHash, _ := idTok.Claim("at_hash").(string); rawAtHash == "" {
		t.Fatal("no at_hash")
	} else if atHash2, err := base64url.DecodeString(rawAtHash); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(atHash2, atHash) {
		t.Error(atHash2)
		t.Fatal(atHash)
	}
}

// POST でないリクエストを拒否できることの検査。
func TestDenyNonPost(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	ta := tadb.New(test_ta.Id()+"a", nil, strsetutil.New(test_rediUri), []jwk.Key{test_taKey}, false, "")
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta, ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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

// 認可コードがおかしかったら拒否できることの検査。
func TestDenyInvalidCode(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id()+"a", hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
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

// 認証リクエストのときと違う redirect_uri を拒否できることの検査。
func TestDenyInvalidRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
		q.Set("redirect_uri", test_rediUri+"a")
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

// 複数のクライアント認証方式が使われていたら拒否できることの検査。
func TestDenyMultiClientAuth(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
		q.Set("client_secret", "abcde")
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

// クライアント認証できなかったら拒否できることの検査。
func TestDenyInvalidClient(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
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
		q.Set("client_assertion", "abcde")
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
	} else if err := "invalid_client"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 期限切れの認可コードを拒否できることの検査。
func TestDenyExpiredCode(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	now := time.Now()
	cod := authcode.New(test_codId, now.Add(time.Nanosecond), test_acntId, now, strsetutil.New("openid"),
		nil, strsetutil.New("email"), test_ta.Id(), test_rediUri, test_nonc)
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

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

// 認可コードが 2 回使われたら拒否できることの検査。
// 2 回使われた認可コードで発行したアクセストークンを無効にできるか。
func TestDenyUsedCode(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt}, []tadb.Element{test_ta})

	cod := newTestCode()
	now := time.Now()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	r, err := newTestRequest(cod.Id(), hndl.selfId+hndl.pathTok)
	if err != nil {
		t.Fatal(err)
	}

	var tokId string
	{
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
			t.Fatal("no token")
		}
		tokId = buff.Access_token
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
	tok, _ := hndl.tokDb.Get(tokId)
	if tok == nil {
		t.Fatal("no token")
	} else if !tok.Invalid() {
		t.Fatal("valid token")
	}
}
