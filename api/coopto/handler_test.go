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

package coopto

import (
	"bytes"
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		"/coop/to",
		20,
		30,
		time.Minute,
		time.Hour,
		time.Minute,
		keydb.NewMemoryDb(keys),
		account.NewMemoryDb(acnts),
		consent.NewMemoryDb(),
		tadb.NewMemoryDb(tas),
		sector.NewMemoryDb(),
		pairwise.NewMemoryDb(),
		coopcode.NewMemoryDb(),
		token.NewMemoryDb(),
		jtidb.NewMemoryDb(),
		rand.New(time.Second),
		true,
	).(*handler)
}

// 1 つ目の ID プロバイダとしての正常系。
// レスポンスが access_token, ids_token を含むことの検査。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
// IDs トークンが署名されていることの検査。
// IDs トークンが iss, sub, aud, exp, iat, ids クレームを含むことの検査。
func TestMainNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
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
		Scope        string
		Ids_token    string
	}
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Access_token == "" {
		t.Fatal("no access token")
	} else if buff.Token_type != "Bearer" {
		t.Error(buff.Token_type)
		t.Fatal("Bearer")
	} else if buff.Expires_in == 0 {
		t.Fatal("no token expiration duration")
	} else if buff.Scope == "" {
		t.Fatal("no scope")
	} else if buff.Ids_token == "" {
		t.Fatal("no IDs token")
	}

	idsTok, err := jwt.Parse([]byte(buff.Ids_token))
	if err != nil {
		t.Fatal(err)
	} else if alg, _ := idsTok.Header("alg").(string); alg != hndl.sigAlg {
		t.Error(alg)
		t.Fatal(hndl.sigAlg)
	} else if !idsTok.IsSigned() {
		t.Fatal("not signed ID token")
	} else if err := idsTok.Verify([]jwk.Key{test_idpKey}); err != nil {
		t.Fatal(err)
	}
	var idsTokBuff struct {
		Iss string
		Sub string
		Aud audience.Audience
		Exp int
		Iat int
		Ids map[string]map[string]interface{}
	}
	if err := json.Unmarshal(idsTok.RawBody(), &idsTokBuff); err != nil {
		t.Fatal(err)
	} else if idsTokBuff.Iss != hndl.selfId {
		t.Error(idsTokBuff.Iss)
		t.Fatal(hndl.selfId)
	} else if idsTokBuff.Sub != test_frTa.Id() {
		t.Error(idsTokBuff.Sub)
		t.Fatal(test_frTa.Id())
	} else if !idsTokBuff.Aud[test_toTa.Id()] {
		t.Error(idsTokBuff.Aud)
		t.Fatal(test_toTa.Id())
	} else if idsTokBuff.Exp == 0 {
		t.Fatal("no exp")
	} else if idsTokBuff.Iat == 0 {
		t.Fatal("no iat")
	} else if idsTokBuff.Iat > idsTokBuff.Exp {
		t.Fatal("iat after exp")
	} else if ids := map[string]map[string]interface{}{
		test_acntTag: {
			"sub": acnt.Id(),
		},
		test_subAcnt1Tag: {
			"sub": subAcnt1.Id(),
		},
	}; !reflect.DeepEqual(idsTokBuff.Ids, ids) {
		t.Error(string(idsTok.RawBody()))
		t.Error(idsTokBuff.Ids)
		t.Fatal(ids)
	}
}

// TA 固有アカウント ID に対応していることの検査。
func TestPairwise(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	toTa := tadb.New("https://to.example.org", nil, nil, []jwk.Key{test_toTaKey}, true, "")
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
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
	}

	var buff struct{ Ids_token string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Ids_token == "" {
		t.Fatal("no IDs token")
	}

	idsTok, err := jwt.Parse([]byte(buff.Ids_token))
	if err != nil {
		t.Fatal(err)
	}
	var idTokBuff struct {
		Ids struct {
			Main struct {
				Sub string
			} `json:"main-user"`
			Sub struct {
				Sub string
			} `json:"sub-user1"`
		}
	}
	if err := json.Unmarshal(idsTok.RawBody(), &idTokBuff); err != nil {
		t.Fatal(err)
	} else if idTokBuff.Ids.Main.Sub == "" || idTokBuff.Ids.Main.Sub == acnt.Id() {
		t.Error("not pairwise")
		t.Fatal(idTokBuff.Ids.Main.Sub)
	} else if idTokBuff.Ids.Sub.Sub == "" || idTokBuff.Ids.Sub.Sub == subAcnt1.Id() {
		t.Error("not pairwise")
		t.Fatal(idTokBuff.Ids.Sub.Sub)
	}
}

// クライアント認証に失敗したら拒否できることの検査。
func TestDenyInvalidTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["client_assertion"] = "abcde"
		body, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
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

// grant_type が無かったら拒否できることの検査。
func TestDenyNoGrantType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testDenyNoSomething(t, "grant_type")
}

// code が無かったら拒否できることの検査。
func TestDenyNoCode(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testDenyNoSomething(t, "code")
}

func testDenyNoSomething(t *testing.T, something string) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		delete(m, something)
		body, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
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

// 仲介コードが有効でなかったら拒否できることの検査。
func TestDenyInvalidCode(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["code"] = test_codId + "a"
		body, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
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

// 仲介コードが発行された連携先でなかったら拒否できることの検査。
func TestDenyDifferentTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	ta := tadb.New("https://dummy.example.org", nil, nil, []jwk.Key{test_toTaKey}, false, "")
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa, ta})

	now := time.Now()
	cod := coopcode.New(test_codId, now.Add(time.Minute), coopcode.NewAccount(test_acntId, test_acntTag),
		test_tokId, test_scop, now.Add(time.Minute/2), []*coopcode.Account{coopcode.NewAccount(test_subAcnt1Id, test_subAcnt1Tag)},
		test_frTa.Id(), ta.Id())
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
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

// user_claims におかしなアカウントタグがあったら拒否できることの検査。
func TestDenyInvalidTag(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["user_claims"] = map[string]map[string]interface{}{
			"unknown": {
				"email": nil,
			},
		}
		body, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
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

// 事前同意が無ければ拒否できることの検査。
func TestDenyNoConsent(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestMainCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{acnt.Id(), subAcnt1.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestMainRequest(hndl.selfId + hndl.pathCoopTo)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["claims"] = map[string]map[string]interface{}{
			"id_token": {
				"email": map[string]interface{}{
					"essential": true,
				},
			},
		}
		body, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Error(w.Code)
		t.Fatal(http.StatusForbidden)
	}
	var buff struct{ Error string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if err := "access_denied"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 2 つ目以降の ID プロバイダとしての正常系。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
// IDs トークンが署名されていることの検査。
// IDs トークンが iss, sub, aud, exp, iat, ids クレームを含むことの検査。
func TestSubNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa})

	now := time.Now()
	cod := newTestSubCode()
	hndl.codDb.Save(cod, now.Add(time.Minute))

	for _, acntId := range []string{subAcnt2.Id()} {
		cons := consent.New(acntId, test_toTa.Id())
		cons.Scope().SetAllow("openid")
		hndl.consDb.Save(cons)
	}

	r, err := newTestSubRequest(hndl.selfId + hndl.pathCoopTo)
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
		Ids_token string
	}
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Ids_token == "" {
		t.Fatal("no IDs token")
	}

	idsTok, err := jwt.Parse([]byte(buff.Ids_token))
	if err != nil {
		t.Fatal(err)
	} else if alg, _ := idsTok.Header("alg").(string); alg != hndl.sigAlg {
		t.Error(alg)
		t.Fatal(hndl.sigAlg)
	} else if !idsTok.IsSigned() {
		t.Fatal("not signed ID token")
	} else if err := idsTok.Verify([]jwk.Key{test_idpKey}); err != nil {
		t.Fatal(err)
	}
	var idsTokBuff struct {
		Iss string
		Sub string
		Aud audience.Audience
		Exp int
		Iat int
		Ids map[string]map[string]interface{}
	}
	if err := json.Unmarshal(idsTok.RawBody(), &idsTokBuff); err != nil {
		t.Fatal(err)
	} else if idsTokBuff.Iss != hndl.selfId {
		t.Error(idsTokBuff.Iss)
		t.Fatal(hndl.selfId)
	} else if idsTokBuff.Sub != test_frTa.Id() {
		t.Error(idsTokBuff.Sub)
		t.Fatal(test_frTa.Id())
	} else if !idsTokBuff.Aud[test_toTa.Id()] {
		t.Error(idsTokBuff.Aud)
		t.Fatal(test_toTa.Id())
	} else if idsTokBuff.Exp == 0 {
		t.Fatal("no exp")
	} else if idsTokBuff.Iat == 0 {
		t.Fatal("no iat")
	} else if idsTokBuff.Iat > idsTokBuff.Exp {
		t.Fatal("iat after exp")
	} else if ids := map[string]map[string]interface{}{
		test_subAcnt2Tag: {
			"sub": subAcnt2.Id(),
		},
	}; !reflect.DeepEqual(idsTokBuff.Ids, ids) {
		t.Error(string(idsTok.RawBody()))
		t.Error(idsTokBuff.Ids)
		t.Fatal(ids)
	}
}
