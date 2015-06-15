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
	"github.com/realglobe-Inc/edo-idp-selector/ticket"
	"github.com/realglobe-Inc/edo-lib/jwk"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

// 正常にログイン UI にリダイレクトさせることの検査。
// セッションが有効ならセッションを発行しないことの検査。
func TestSelectPage(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(page.ticExpIn)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("POST", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(uri)
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	} else if ok, err := regexp.MatchString(page.sessLabel+"=[0-9a-zA-Z_\\-]", w.HeaderMap.Get("Set-Cookie")); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Error("new session")
		t.Fatal(w.HeaderMap.Get("Set-Cookie"))
	}
}

// セッションが無ければセッションを発行することの検査。
func TestSelectPageSessionPublication(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})

	r, err := http.NewRequest("Post", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	page.HandleAuth(w, r)

	if ok, err := regexp.MatchString(page.sessLabel+"=[0-9a-zA-Z_\\-]", w.HeaderMap.Get("Set-Cookie")); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Error("no new session")
		t.Fatal(w.HeaderMap.Get("Set-Cookie"))
	}
}

// 同意済みセッションならクライアントにリダイレクトさせることの検査。
func TestSelectPageRedirectClient(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	acnt := session.NewAccount(test_acnt.Id(), test_acnt.Name())
	acnt.Login()
	sess.SelectAccount(acnt)
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(page.ticExpIn)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	cons := consent.New(test_acnt.Id(), test_ta.Id())
	cons.Scope().SetAllow("openid")
	page.consDb.Save(cons)

	r, err := http.NewRequest("POST", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if rediUri := uri.Scheme + "://" + uri.Host + uri.Path; rediUri != test_rediUri {
		t.Error(w.HeaderMap.Get("Location"))
		t.Error(rediUri)
		t.Fatal(test_rediUri)
	}
}

// 知らないパラメータを無視できることの検査。
func TestSelectPageIgnoreUnknownParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(page.ticExpIn)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("POST", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Set("aaaa", "bbbb")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

	if w.Code != http.StatusFound {
		t.Error(w.Code)
		t.Fatal(http.StatusFound)
	} else if uri, err := url.Parse(w.HeaderMap.Get("Location")); err != nil {
		t.Fatal(err)
	} else if uri.Path != page.pathLginUi {
		t.Error(w.HeaderMap.Get("Location"))
		t.Error(uri.Path)
		t.Fatal(page.pathLginUi)
	}
}

// 重複パラメータを拒否できることの検査。
func TestSelectPageDenyOverlapParameter(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(page.ticExpIn)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("POST", page.selfId+"/select?"+test_selQuery+"&ticket=aaaa", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

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

// 入力券が無ければ拒否できることの検査。
func TestSelectPageDenyNoTicket(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(page.ticExpIn)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("POST", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Del("ticket")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

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

// 入力券がおかしければ拒否できることの検査。
func TestSelectPageDenyInvalidTicket(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(page.ticExpIn)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("POST", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})
	{
		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		q, err := url.ParseQuery(string(buff))
		if err != nil {
			t.Fatal(err)
		}
		q.Set("ticket", test_ticId+"a")
		r.Body = ioutil.NopCloser(strings.NewReader(q.Encode()))
	}

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

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

// 入力券が期限切れなら拒否できることの検査。
func TestSelectPageDenyExpiredTicket(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	page := newTestPage([]jwk.Key{test_idpKey}, nil, []account.Element{test_acnt}, []tadb.Element{test_ta})
	now := time.Now()
	sess := session.New(test_sessId, now.Add(page.sessExpIn))
	sess.SetRequest(test_req)
	sess.SetTicket(ticket.New(test_ticId, now.Add(-1)))
	page.sessDb.Save(sess, now.Add(page.sessDbExpIn))

	r, err := http.NewRequest("POST", page.selfId+"/select", strings.NewReader(test_selQuery))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{
		Name:  page.sessLabel,
		Value: sess.Id(),
	})

	w := httptest.NewRecorder()
	page.HandleSelect(w, r)

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
