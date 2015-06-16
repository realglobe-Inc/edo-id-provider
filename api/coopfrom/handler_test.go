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

package coopfrom

import (
	"bytes"
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
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

func newTestHandler(keys []jwk.Key, acnts []account.Element, tas []tadb.Element, idps []idpdb.Element) *handler {
	return New(
		server.NewStopper(),
		"https://idp.example.org",
		"ES256",
		"",
		"/coop/from",
		30,
		time.Minute,
		time.Hour,
		20,
		time.Minute,
		keydb.NewMemoryDb(keys),
		pairwise.NewMemoryDb(),
		account.NewMemoryDb(acnts),
		tadb.NewMemoryDb(tas),
		idpdb.NewMemoryDb(idps),
		coopcode.NewMemoryDb(),
		token.NewMemoryDb(),
		jtidb.NewMemoryDb(),
		rand.New(time.Second),
		true,
	).(*handler)
}

// ID プロバイダが 1 つの場合の正常系。
// レスポンスが code_token を含むことの検査。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
// コードトークンが署名されていることの検査。
// コードトークンが iss, sub, aud, from_client, user_tag, user_tags クレームを含むことの検査。
func TestSingleNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
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
		Code_token string
	}
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Code_token == "" {
		t.Fatal("no code token")
	}

	codTok, err := jwt.Parse([]byte(buff.Code_token))
	if err != nil {
		t.Fatal(err)
	} else if alg, _ := codTok.Header("alg").(string); alg != hndl.sigAlg {
		t.Error(alg)
		t.Fatal(hndl.sigAlg)
	} else if !codTok.IsSigned() {
		t.Fatal("not signed ID token")
	} else if err := codTok.Verify([]jwk.Key{test_idpKey}); err != nil {
		t.Fatal(err)
	}
	var codTokBuff struct {
		Iss         string
		Sub         string
		Aud         audience.Audience
		From_client string
		User_tag    string
		User_tags   strset.Set
	}
	if err := json.Unmarshal(codTok.RawBody(), &codTokBuff); err != nil {
		t.Fatal(err)
	} else if codTokBuff.Iss != hndl.selfId {
		t.Error(codTokBuff.Iss)
		t.Fatal(hndl.selfId)
	} else if codTokBuff.Sub == "" {
		t.Fatal("no code")
	} else if !codTokBuff.Aud[test_toTa.Id()] {
		t.Error(codTokBuff.Aud)
		t.Fatal(test_toTa.Id())
	} else if codTokBuff.From_client != test_frTa.Id() {
		t.Error(codTokBuff.From_client)
		t.Fatal(test_frTa.Id())
	} else if codTokBuff.User_tag != test_acntTag {
		t.Error(codTokBuff.User_tag)
		t.Fatal(test_acntTag)
	} else if acntTags := strsetutil.New(test_subAcnt1Tag); !reflect.DeepEqual(map[string]bool(codTokBuff.User_tags), acntTags) {
		t.Error(string(codTok.RawBody()))
		t.Error(codTokBuff.User_tags)
		t.Fatal(acntTags)
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
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
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

// response_type が無かったら拒否できることの検査。
func TestDenyNoResponseType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "response_type")
}

// grant_type が無かったら拒否できることの検査。
func TestDenyNoGrantType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "grant_type")
}

// 主体の ID プロバイダの場合に、from_client が無かったら拒否できることの検査。
func TestMainDenyNoFromTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "from_client")
}

// 主体の ID プロバイダの場合に、to_client が無かったら拒否できることの検査。
func TestMainDenyNoToTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "to_client")
}

// 主体の ID プロバイダの場合に、access_token が無かったら拒否できることの検査。
func TestMainDenyNoAccessToken(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "access_token")
}

// 主体の ID プロバイダの場合に、user_tag が無かったら拒否できることの検査。
func TestMainDenyNoUserTag(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "user_tag")
}

func testMainDenyNoSomething(t *testing.T, something string) {
	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
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

// 主体の ID プロバイダの場合に、アクセストークンが有効でなかったら拒否できることの検査。
func TestMainDenyInvalidAccessToken(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
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

// 主体の ID プロバイダの場合に、許されない scope を含むなら拒否できることの検査。
func TestMainDenyInvalidScope(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["scope"] = "openid phone"
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
	} else if err := "invalid_scope"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 主体の ID プロバイダの場合に、要請先 TA が存在しないなら拒否できることの検査。
func TestMainDenyInvalidToTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["to_client"] = test_toTa.Id() + "a"
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

// 主体の ID プロバイダの場合に、要請元 TA と要請先 TA が同じなら拒否できることの検査。
func TestMainDenySameTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["to_client"] = test_frTa.Id()
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

// users におかしなユーザーいたら拒否できることの検査。
func TestDenyInvalidUsers(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["users"] = map[string]string{test_subAcnt1Tag: subAcnt1.Id() + "a"}
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

// ユーザータグに重複があったら拒否できることの検査。
func TestDenyTagOverlap(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequest(hndl.selfId + hndl.pathCoopFr)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["users"] = map[string]string{test_acntTag: subAcnt1.Id()}
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
