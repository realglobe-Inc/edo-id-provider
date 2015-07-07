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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/rglog/level"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
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
		rand.New(time.Minute),
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{"client_assertion": "abcde"}, nil)
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
	} else if err := "invalid_client"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// response_type が無かったら拒否できることの検査。
func TestDenyNoResponseType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "response_type")
}

// grant_type が無かったら拒否できることの検査。
func TestDenyNoGrantType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "grant_type")
}

// 主体の ID プロバイダの場合に、from_client が無かったら拒否できることの検査。
func TestMainDenyNoFromTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "from_client")
}

// 主体の ID プロバイダの場合に、to_client が無かったら拒否できることの検査。
func TestMainDenyNoToTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "to_client")
}

// 主体の ID プロバイダの場合に、access_token が無かったら拒否できることの検査。
func TestMainDenyNoAccessToken(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testMainDenyNoSomething(t, "access_token")
}

// 主体の ID プロバイダの場合に、user_tag が無かったら拒否できることの検査。
func TestMainDenyNoUserTag(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{something: nil}, nil)
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 主体の ID プロバイダの場合に、アクセストークンが有効でなかったら拒否できることの検査。
func TestMainDenyInvalidAccessToken(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
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
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{"scope": "openid phone"}, nil)
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
	} else if err := "invalid_scope"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 主体の ID プロバイダの場合に、連携先 TA が存在しないなら拒否できることの検査。
func TestMainDenyInvalidToTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{"to_client": test_toTa.Id() + "a"}, nil)
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 主体の ID プロバイダの場合に、連携元 TA と連携先 TA が同じなら拒否できることの検査。
func TestMainDenySameTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{"to_client": test_frTa.Id()}, nil)
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// users におかしなユーザーいたら拒否できることの検査。
func TestDenyInvalidUsers(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{"users": map[string]string{test_subAcnt1Tag: subAcnt1.Id() + "a"}}, nil)
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// ユーザータグに重複があったら拒否できることの検査。
func TestDenyTagOverlap(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, nil)

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestSingleRequestWithParams(hndl.selfId+hndl.pathCoopFr, map[string]interface{}{"users": map[string]string{test_acntTag: subAcnt1.Id()}}, nil)
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// ID プロバイダが 2 以上ある 1 つ目の場合の正常系。
// レスポンスが code_token, referral を含むことの検査。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
// コードトークンが署名されていることの検査。
// コードトークンが iss, sub, aud, from_client, user_tag, user_tags クレームを含むことの検査。
// referral が署名されていることの検査。
// referral が iss, sub, aud, exp, jti, to_client, related_users, hash_alg クレームを含むことの検査。
func TestMainNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestMainRequest(hndl.selfId+hndl.pathCoopFr, test_idp2.Id())
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
		Referral   string
	}
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Code_token == "" {
		t.Fatal("no code token")
	} else if buff.Referral == "" {
		t.Fatal("no referral")
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
	var refHash []byte
	ref, err := jwt.Parse([]byte(buff.Referral))
	if err != nil {
		t.Fatal(err)
	} else if alg, _ := ref.Header("alg").(string); alg != hndl.sigAlg {
		t.Error(alg)
		t.Fatal(hndl.sigAlg)
	} else if !ref.IsSigned() {
		t.Fatal("not signed ID referral")
	} else if err := ref.Verify([]jwk.Key{test_idpKey}); err != nil {
		t.Fatal(err)
	} else if hGen := jwt.HashGenerator(alg); !hGen.Available() {
		t.Error(hGen)
		t.Fatal("unsupported algorithm " + alg)
	} else {
		hFun := hGen.New()
		hFun.Write([]byte(buff.Referral))
		hVal := hFun.Sum(nil)
		refHash = hVal[:len(hVal)/2]
	}

	var codTokBuff struct {
		Iss         string
		Sub         string
		Aud         audience.Audience
		From_client string
		User_tag    string
		User_tags   strset.Set
		Ref_Hash    string
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
	} else if refHash2, err := base64url.DecodeString(codTokBuff.Ref_Hash); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(refHash2, refHash) {
		t.Error(refHash2)
		t.Fatal(refHash)
	}
	var refBuff struct {
		Iss           string
		Sub           string
		Aud           audience.Audience
		Exp           int64
		Jti           string
		To_client     string
		Related_users map[string]string
		Hash_alg      string
	}
	if err := json.Unmarshal(ref.RawBody(), &refBuff); err != nil {
		t.Fatal(err)
	} else if refBuff.Iss != hndl.selfId {
		t.Error(refBuff.Iss)
		t.Fatal(hndl.selfId)
	} else if refBuff.Sub != test_frTa.Id() {
		t.Error(refBuff.Sub)
		t.Fatal(test_frTa.Id())
	} else if aud := audience.New(test_idp2.Id()); !reflect.DeepEqual(refBuff.Aud, aud) {
		t.Error(refBuff.Aud)
		t.Fatal(aud)
	} else if exp := time.Unix(refBuff.Exp, 0); time.Now().After(exp) {
		t.Error("expired")
		t.Fatal(exp)
	} else if refBuff.Jti == "" {
		t.Fatal("no JWT ID")
	} else if refBuff.To_client != test_toTa.Id() {
		t.Error(refBuff.To_client)
		t.Fatal(test_toTa.Id())
	} else if relAcnts := map[string]string{test_subAcnt2Tag: calcTestSubAccount2HashValue(test_idp2.Id())}; !reflect.DeepEqual(refBuff.Related_users, relAcnts) {
		t.Error(refBuff.Related_users)
		t.Fatal(relAcnts)
	} else if refBuff.Hash_alg != test_hAlg {
		t.Error(refBuff.Hash_alg)
		t.Fatal(test_hAlg)
	}
}

// ID プロバイダが 2 以上ある 2 つ目以降の場合の正常系。
// レスポンスが code_token を含むことの検査。
// レスポンスが Cache-Control: no-store, Pragma: no-cache ヘッダを含むことの検査。
// コードトークンが署名されていることの検査。
// コードトークンが iss, sub, aud, user_tags, ref_hash クレームを含むことの検査。
func TestSubNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, refHash, err := newTestSubRequest(hndl.selfId, hndl.selfId+hndl.pathCoopFr)
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

	var buff struct{ Code_token string }
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
		Iss       string
		Sub       string
		Aud       audience.Audience
		User_tags strset.Set
		Ref_Hash  string
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
	} else if acntTags := strsetutil.New(test_subAcnt2Tag); !reflect.DeepEqual(map[string]bool(codTokBuff.User_tags), acntTags) {
		t.Error(codTokBuff.User_tags)
		t.Fatal(acntTags)
	} else if refHash2, err := base64url.DecodeString(codTokBuff.Ref_Hash); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(refHash2, refHash) {
		t.Error(refHash2)
		t.Fatal(refHash)
	}
}

// 2 つ目以降の ID プロバイダで TA 固有のアカウント ID に対応できることの検査。
func TestPairwise(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	frTa := tadb.New(test_frTa.Id(), nil, nil, test_frTa.Keys(), true, "")
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{frTa, test_toTa}, []idpdb.Element{test_idp2})

	pw := pairwise.Generate(subAcnt2.Id(), frTa.Sector(), []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19})
	hndl.pwDb.Save(pw)

	r, refHash, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, map[string]interface{}{
		"users": map[string]string{test_subAcnt2Tag: pw.Pairwise()},
	}, map[string]interface{}{
		"related_users": map[string]string{test_subAcnt2Tag: calcTestAccountHashValue(hndl.selfId, pw.Pairwise())},
	}, nil)
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

	var buff struct{ Code_token string }
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
		Iss       string
		Sub       string
		Aud       audience.Audience
		User_tags strset.Set
		Ref_Hash  string
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
	} else if acntTags := strsetutil.New(test_subAcnt2Tag); !reflect.DeepEqual(map[string]bool(codTokBuff.User_tags), acntTags) {
		t.Error(codTokBuff.User_tags)
		t.Fatal(acntTags)
	} else if refHash2, err := base64url.DecodeString(codTokBuff.Ref_Hash); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(refHash2, refHash) {
		t.Error(refHash2)
		t.Fatal(refHash)
	}
}

// related_issuers におかしな ID プロバイダが含まれるなら拒否できることの検査。
func TestDenyInvalidIdProvider(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestMainAccount()
	subAcnt1 := newTestSubAccount1()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{acnt, subAcnt1}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_frTa.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := newTestMainRequest(hndl.selfId+hndl.pathCoopFr, test_idp2.Id()+"a")
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// referral の署名がおかしかったら拒否できることの検査。
func TestSubDenyInvalidReferral(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequest(hndl.selfId, hndl.selfId+hndl.pathCoopFr)
	if err != nil {
		t.Fatal(err)
	}
	{
		var m map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		m["referral"] = regexp.MustCompile("\\.[^.]+$").ReplaceAllString(m["referral"].(string), ".AAAA")
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

// 主体でないアカウントの ID プロバイダの場合に、response_type が無かったら拒否できることの検査。
func TestSubDenyNoResponseType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyNoSomething(t, "response_type")
}

// 主体でないアカウントの ID プロバイダの場合に、grant_type が無かったら拒否できることの検査。
func TestSubDenyNoGrantType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyNoSomething(t, "grant_type")
}

// 主体でないアカウントの ID プロバイダの場合に、referral が無かったら拒否できることの検査。
func TestSubDenyNoReferral(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyNoSomething(t, "referral")
}

// 主体でないアカウントの ID プロバイダの場合に、users が無かったら拒否できることの検査。
func TestSubDenyNoUsers(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyNoSomething(t, "users")
}

func testSubDenyNoSomething(t *testing.T, something string) {
	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, map[string]interface{}{something: nil}, nil, nil)
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
	} else if err := "invalid_request"; buff.Error != err {
		t.Error(buff.Error)
		t.Fatal(err)
	}
}

// 主体でないアカウントの ID プロバイダの場合に、referral に iss が無かったら拒否できることの検査。
func TestSubDenyReferralNoIss(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "iss")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に sub が無かったら拒否できることの検査。
func TestSubDenyReferralNoSub(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "sub")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に aud が無かったら拒否できることの検査。
func TestSubDenyReferralNoAud(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "aud")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に exp が無かったら拒否できることの検査。
func TestSubDenyReferralNoExp(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "exp")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に jti が無かったら拒否できることの検査。
func TestSubDenyReferralNoJti(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "jti")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に to_client が無かったら拒否できることの検査。
func TestSubDenyReferralNoToClient(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "to_client")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に related_users が無かったら拒否できることの検査。
func TestSubDenyReferralNoRelatedUsers(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "related_users")
}

// 主体でないアカウントの ID プロバイダの場合に、referral に hash_alg が無かったら拒否できることの検査。
func TestSubDenyReferralNoHashAlg(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	testSubDenyReferralNoSomething(t, "hash_alg")
}

func testSubDenyReferralNoSomething(t *testing.T, something string) {
	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, nil, map[string]interface{}{something: nil}, nil)
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

// 主体でないアカウントの ID プロバイダの場合に、連携先 TA が存在しないなら拒否できることの検査。
func TestSubDenyInvalidToTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, nil, map[string]interface{}{"to_client": test_toTa.Id() + "a"}, nil)
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

// 主体でないアカウントの ID プロバイダの場合に、連携元 TA と連携先 TA が同じなら拒否できることの検査。
func TestSubDenySameTa(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, nil, map[string]interface{}{"to_client": test_frTa.Id()}, nil)
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

// 主体でないアカウントの ID プロバイダの場合に、related_users に users のユーザーがなかったら拒否できることの検査。
func TestDenyInvalidNoRelatedUsers(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, nil, map[string]interface{}{"related_users": map[string]string{test_subAcnt2Tag + "a": calcTestSubAccount2HashValue(test_idp2.Id())}}, nil)
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

// related_users におかしなユーザーいたら拒否できることの検査。
func TestDenyInvalidRelatedUsers(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	subAcnt2 := newTestSubAccount2()
	hndl := newTestHandler([]jwk.Key{test_idpKey}, []account.Element{subAcnt2}, []tadb.Element{test_frTa, test_toTa}, []idpdb.Element{test_idp2})

	r, _, err := newTestSubRequestWithParams(hndl.selfId, hndl.selfId+hndl.pathCoopFr, nil, map[string]interface{}{"related_users": map[string]string{test_subAcnt2Tag: calcTestSubAccount2HashValue(test_idp2.Id()) + "a"}}, nil)
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
