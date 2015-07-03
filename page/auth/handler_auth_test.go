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

package auth

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

// 正常にログイン UI にリダイレクトさせることの検査。
// セッションが無ければセッションを発行することの検査。
func TestAuthPage(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	} else if ok, err := regexp.MatchString(page.sessLabel+"=[0-9a-zA-Z_\\-]", w.HeaderMap.Get("Set-Cookie")); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Error("no new session")
		t.Fatal(w.HeaderMap.Get("Set-Cookie"))
	}
}

// セッションが有効ならセッションを発行しないことの検査。
func TestAuthPageNoSessionPublication(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if ok, err := regexp.MatchString(page.sessLabel+"=[0-9a-zA-Z_\\-]", w.HeaderMap.Get("Set-Cookie")); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Error("new session")
		t.Fatal(w.HeaderMap.Get("Set-Cookie"))
	} else if sess, _ = page.sessDb.Get(sess.Id()); sess == nil {
		t.Fatal("no session")
	} else if req := sess.Request(); req == nil {
		t.Fatal("no request")
	} else if respType := strsetutil.New("code"); !reflect.DeepEqual(req.ResponseType(), respType) {
		t.Error(req.ResponseType())
		t.Fatal(respType)
	} else if scop := strsetutil.New("openid"); !reflect.DeepEqual(req.Scope(), scop) {
		t.Error(req.Scope())
		t.Fatal(scop)
	} else if req.Ta() != test_ta.Id() {
		t.Error(req.Ta())
		t.Fatal(test_ta.Id())
	} else if req.RedirectUri() != test_rediUri {
		t.Error(req.RedirectUri())
		t.Fatal(test_rediUri)
	} else if req.State() != test_stat {
		t.Error(req.State())
		t.Fatal(test_stat)
	} else if req.Nonce() != test_nonc {
		t.Error(req.Nonce())
		t.Fatal(test_nonc)
	}
}

// セッションの期限が切れそうならセッションを更新することの検査。
func TestAuthPageSessionRefresh(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessRefDelay-time.Nanosecond))
	sess.SelectAccount(session.NewAccount(test_acnt.Id(), test_acnt.Name()))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if ok, err := regexp.MatchString(page.sessLabel+"=[0-9a-zA-Z_\\-]", w.HeaderMap.Get("Set-Cookie")); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Error("no new session")
		t.Fatal(w.HeaderMap.Get("Set-Cookie"))
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri.Path)
		t.Fatal(page.pathSelUi)
	} else if query := uri.Query(); len(query) == 0 {
		t.Fatal("no query")
	} else if acnts, acnts2 := `["`+test_acnt.Name()+`"]`, query.Get("usernames"); acnts2 != acnts {
		t.Error(acnts2)
		t.Fatal(acnts)
	}
}

// 認証・同意済みセッションならクライアントにリダイレクトさせることの検査。
func TestAuthPageRedirectClient(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if rediUri := uri.Scheme + "://" + uri.Host + uri.Path; rediUri != test_rediUri {
		t.Error(w.HeaderMap.Get("Location"))
		t.Error(rediUri)
		t.Fatal(test_rediUri)
	} else if q := uri.Query(); len(q) == 0 {
		t.Fatal("no query")
	} else if q.Get("code") == "" {
		t.Fatal("no code")
	}
}

// 知らないパラメータを無視できることの検査。
func TestAuthPageIgnoreUnknownParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery+"&aaaa=bbbb", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if sess, _ = page.sessDb.Get(sess.Id()); sess == nil {
		t.Fatal("no session")
	} else if req := sess.Request(); req == nil {
		t.Fatal("no request")
	}
}

// 重複パラメータを拒否できることの検査。
func TestAuthPageDenyOverlapParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery+"&nonce=aaaa", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "invalid_request", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// scope が無かったら拒否できることの検査。
func TestAuthPageDenyNoScope(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Del("scope")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "invalid_request", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// 知らない scope を無視できることの検査。
func TestAuthPageIgnoreUnknownScope(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		scop := request.FormValueSet(q.Get("scope"))
		scop["unknown"] = true
		q.Set("scope", request.ValueSetForm(scop))
		r.URL.RawQuery = q.Encode()
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if sess, _ = page.sessDb.Get(sess.Id()); sess == nil {
		t.Fatal("no session")
	} else if req := sess.Request(); req == nil {
		t.Fatal("no request")
	}
}

// client_id が無かったら拒否できることの検査。
func TestAuthPageDenyNoClientId(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Del("client_id")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	}
}

// response_type が無かったら拒否できることの検査。
func TestAuthPageDenyNoResponseType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Del("response_type")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "invalid_request", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// 知らない response_type を拒否できることの検査。
func TestAuthPageDenyUnknownResponseType(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		respType := request.FormValueSet(q.Get("response_type"))
		respType["unknonw"] = true
		q.Set("response_type", request.ValueSetForm(respType))
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "unsupported_response_type", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// エラーリダイレクトで redirect_uri のパラメータを維持できることの検査。
func TestAuthPageKeepRedirectParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	ta := tadb.New(
		test_ta.Id(),
		test_ta.Names(),
		strsetutil.New(test_rediUri+"?aaaa=bbbb"),
		test_ta.Keys(),
		test_ta.Pairwise(),
		test_ta.Sector(),
	)
	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Set("redirect_uri", test_rediUri+"?aaaa=bbbb")
		q.Del("response_type")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if v := q.Get("aaaa"); v != "bbbb" {
		t.Error(v)
		t.Fatal("bbbb")
	}
}

// redirect_uri がおかしかったら拒否できることの検査。
func TestAuthPageDenyInvalidRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Set("redirect_uri", test_rediUri+"a")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	}
}

// redirect_uri が無かったら拒否できることの検査。
func TestAuthPageDenyNoRedirectUri(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Del("redirect_uri")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusBadRequest {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	}
}

// エラーリダイレクトで state パラメータを返せることの検査。
func TestAuthPageReturnState(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	{
		q := r.URL.Query()
		q.Del("response_type")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if stat := q.Get("state"); stat != test_stat {
		t.Error(stat)
		t.Fatal(test_stat)
	}
}

// POST でも正常にログイン UI にリダイレクトさせることの検査。
func TestAuthPagePost(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("POST", page.selfId+"/auth", strings.NewReader(test_authQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	}
}

// prompt が login を含むならログイン UI にリダイレクトさせることの検査。
func TestAuthPageForceLogin(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "login")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	}
}

// prompt が none と login を含むなら拒否できることの検査。
func TestAuthPageDenyLoginWithoutUi(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "login none")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "login_required", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// prompt が consent を含むならログイン UI にリダイレクトさせることの検査。
func TestAuthPageForceConsent(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "consent")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathConsUi {
		t.Error(uri.Path)
		t.Fatal(page.pathConsUi)
	}
}

// prompt が none と consent を含むなら拒否できることの検査。
func TestAuthPageDenyConsentWithoutUi(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "consent none")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "consent_required", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// prompt が select を含むならログイン UI にリダイレクトさせることの検査。
func TestAuthPageForceSelect(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "select_account")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathSelUi {
		t.Error(uri.Path)
		t.Fatal(page.pathSelUi)
	}
}

// prompt が none と select を含むなら拒否できることの検査。
func TestAuthPageDenySelectWithoutUi(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "select_account none")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "account_selection_required", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// max_age より時間が経っているならログイン UI にリダイレクトさせることの検査。
func TestAuthPageRedirectLoginUiIfExpired(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("max_age", "0")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	}
}

// アカウント選択 UI にパラメータを渡せることの検査。
func TestAuthPageSelectUiParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SelectAccount(session.NewAccount(test_acnt.Id(), test_acnt.Name()))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("prompt", "select_account")
		q.Set("display", "page")
		q.Set("ui_locales", "ja-JP en-US")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathSelUi {
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if idp := q.Get("issuer"); idp != page.selfId {
		t.Error(idp)
		t.Fatal(page.selfId)
	} else if acntNames, acntNames2 := `["`+test_acnt.Name()+`"]`, q.Get("usernames"); acntNames2 != acntNames {
		t.Error(acntNames2)
		t.Fatal(acntNames)
	} else if disp := q.Get("display"); disp != "page" {
		t.Error(disp)
		t.Fatal("page")
	} else if langs := q.Get("locales"); langs != "ja-JP en-US" {
		t.Error(langs)
		t.Fatal("ja-JP en-US")
	} else if msg := q.Get("message"); msg == "" {
		t.Error("no message")
	}
}

// ログイン UI にパラメータを渡せることの検査。
func TestAuthPageLoginUiParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SelectAccount(session.NewAccount(test_acnt.Id(), test_acnt.Name()))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("display", "page")
		q.Set("ui_locales", "ja-JP en-US")
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if idp := q.Get("issuer"); idp != page.selfId {
		t.Error(idp)
		t.Fatal(page.selfId)
	} else if acntNames, acntNames2 := `["`+test_acnt.Name()+`"]`, q.Get("usernames"); acntNames2 != acntNames {
		t.Error(acntNames2)
		t.Fatal(acntNames)
	} else if disp := q.Get("display"); disp != "page" {
		t.Error(disp)
		t.Fatal("page")
	} else if langs := q.Get("locales"); langs != "ja-JP en-US" {
		t.Error(langs)
		t.Fatal("ja-JP en-US")
	} else if msg := q.Get("message"); msg == "" {
		t.Error("no message")
	}
}

// 同意 UI にパラメータを渡せることの検査。
func TestAuthPageConsentUiParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("display", "page")
		q.Set("ui_locales", "ja-JP en-US")
		q.Set("claims", `{"userinfo":{"pds":{"essential":true},"email":null}}`)
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathConsUi {
		t.Error(uri.Path)
		t.Fatal(page.pathConsUi)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if idp := q.Get("issuer"); idp != page.selfId {
		t.Error(idp)
		t.Fatal(page.selfId)
	} else if acntName := q.Get("username"); acntName != test_acnt.Name() {
		t.Error(acntName)
		t.Fatal(test_acnt.Name())
	} else if scop := q.Get("scope"); scop != "openid" {
		t.Error(scop)
		t.Fatal("openid")
	} else if clms := q.Get("claims"); clms != "pds" {
		t.Error(clms)
		t.Fatal("pds")
	} else if optClms := q.Get("optional_claims"); optClms != "email" {
		t.Error(optClms)
		t.Fatal("email")
	} else if exp, exp2 := strconv.Itoa(int(page.tokExpIn/time.Second)), q.Get("expires_in"); exp2 != exp {
		t.Error(exp2)
		t.Fatal(exp)
	} else if taId := q.Get("client_id"); taId != test_ta.Id() {
		t.Error(taId)
		t.Fatal(test_ta.Id())
	} else if disp := q.Get("display"); disp != "page" {
		t.Error(disp)
		t.Fatal("page")
	} else if langs := q.Get("locales"); langs != "ja-JP en-US" {
		t.Error(langs)
		t.Fatal("ja-JP en-US")
	} else if msg := q.Get("message"); msg == "" {
		t.Error("no message")
	}
}

// claims を受け取れることの検査。
func TestAuthPageClaims(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("claims", `{"userinfo":{"pds":{"essential":true},"email":null,"sub":{"value":"`+test_acnt.Id()+`"}}}`)
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if sess, _ = page.sessDb.Get(sess.Id()); sess == nil {
		t.Fatal("no session")
	} else if req := sess.Request(); req == nil {
		t.Fatal("no request")
	} else if clms := req.Claims().AccountEntries(); len(clms) == 0 {
		t.Fatal("no claims")
	}
}

// 必須クレームが無かったら拒否できることの検査。
func TestAuthDenyAbsentRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("claims", `{"userinfo":{"something":{"essential":true}}}`)
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "access_denied", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// 違う値が要求されたら拒否できることの検査。
func TestAuthDenyInvalidClaim(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("claims", `{"userinfo":{"sub":{"value":"`+test_acnt.Id()+`a"}}}`)
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if q := uri.Query(); q == nil {
		t.Fatal("no parameter")
	} else if err, err2 := "access_denied", q.Get("error"); err2 != err {
		t.Error(err2)
		t.Fatal(err)
	}
}

// request を受け取れることの検査。
func TestAuthRequestParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	jt := jwt.New()
	jt.SetHeader("alg", "none")
	reqQ, err := url.ParseQuery(test_authQuery)
	if err != nil {
		t.Fatal(err)
	}
	for k, vs := range reqQ {
		jt.SetClaim(k, vs[0])
	}
	req, err := jt.Encode()
	if err != nil {
		t.Fatal(err)
	}

	q := url.Values{}
	q.Set("response_type", reqQ.Get("response_type"))
	q.Set("client_id", reqQ.Get("client_id"))
	q.Set("request", string(req))
	r, err := http.NewRequest("GET", page.selfId+"/auth?"+q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if sess, _ = page.sessDb.Get(sess.Id()); sess == nil {
		t.Fatal("no session")
	} else if req := sess.Request(); req == nil {
		t.Fatal("no request")
	} else if respType := strsetutil.New("code"); !reflect.DeepEqual(req.ResponseType(), respType) {
		t.Error(req.ResponseType())
		t.Fatal(respType)
	} else if scop := strsetutil.New("openid"); !reflect.DeepEqual(req.Scope(), scop) {
		t.Error(req.Scope())
		t.Fatal(scop)
	} else if req.Ta() != test_ta.Id() {
		t.Error(req.Ta())
		t.Fatal(test_ta.Id())
	} else if req.RedirectUri() != test_rediUri {
		t.Error(req.RedirectUri())
		t.Fatal(test_rediUri)
	} else if req.State() != test_stat {
		t.Error(req.State())
		t.Fatal(test_stat)
	} else if req.Nonce() != test_nonc {
		t.Error(req.Nonce())
		t.Fatal(test_nonc)
	}
}

// request_uri を受け取れることの検査。
func TestAuthRequestUriParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	reqUri := "https://ta.example.org/request"

	jt := jwt.New()
	jt.SetHeader("alg", "none")
	reqQ, err := url.ParseQuery(test_authQuery)
	if err != nil {
		t.Fatal(err)
	}
	for k, vs := range reqQ {
		jt.SetClaim(k, vs[0])
	}
	req, err := jt.Encode()
	if err != nil {
		t.Fatal(err)
	}
	page := newTestPage([]jwk.Key{test_idpKey}, []webdb.Element{webdb.New(reqUri, req)}, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	q := url.Values{}
	q.Set("response_type", reqQ.Get("response_type"))
	q.Set("client_id", reqQ.Get("client_id"))
	q.Set("request_uri", reqUri)
	r, err := http.NewRequest("GET", page.selfId+"/auth?"+q.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if sess, _ = page.sessDb.Get(sess.Id()); sess == nil {
		t.Fatal("no session")
	} else if req := sess.Request(); req == nil {
		t.Fatal("no request")
	} else if respType := strsetutil.New("code"); !reflect.DeepEqual(req.ResponseType(), respType) {
		t.Error(req.ResponseType())
		t.Fatal(respType)
	} else if scop := strsetutil.New("openid"); !reflect.DeepEqual(req.Scope(), scop) {
		t.Error(req.Scope())
		t.Fatal(scop)
	} else if req.Ta() != test_ta.Id() {
		t.Error(req.Ta())
		t.Fatal(test_ta.Id())
	} else if req.RedirectUri() != test_rediUri {
		t.Error(req.RedirectUri())
		t.Fatal(test_rediUri)
	} else if req.State() != test_stat {
		t.Error(req.State())
		t.Fatal(test_stat)
	} else if req.Nonce() != test_nonc {
		t.Error(req.Nonce())
		t.Fatal(test_nonc)
	}
}

// response_type が code id_token で動作することの検査。
func TestAuthIdToken(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("GET", page.selfId+"/auth?"+test_authQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		q := r.URL.Query()
		q.Set("response_type", request.ValueSetForm(strsetutil.New("code", "id_token")))
		r.URL.RawQuery = q.Encode()
	}

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if rediUri := uri.Scheme + "://" + uri.Host + uri.Path; rediUri != test_rediUri {
		t.Error(w.HeaderMap.Get("Location"))
		t.Error(rediUri)
		t.Fatal(test_rediUri)
	} else if q := uri.Query(); len(q) == 0 {
		t.Fatal("no query")
	} else if q.Get("code") == "" {
		t.Fatal("no code")
	} else if rawIdTok := q.Get("id_token"); rawIdTok == "" {
		t.Fatal("no ID token")
	} else if idTok, err := jwt.Parse([]byte(rawIdTok)); err != nil {
		t.Fatal(err)
	} else if alg, _ := idTok.Header("alg").(string); alg != page.sigAlg {
		t.Error(alg)
		t.Fatal(page.sigAlg)
	} else if !idTok.IsSigned() {
		t.Fatal("not signed ID token")
	} else if err := idTok.Verify([]jwk.Key{test_idpKey}); err != nil {
		t.Fatal(err)
	} else if iss, _ := idTok.Claim("iss").(string); iss != page.selfId {
		t.Error(iss)
		t.Fatal(page.selfId)
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
	}
}
