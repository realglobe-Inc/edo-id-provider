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

package account

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole(logRoot, level.OFF)
}

func newTestHandler(acnts []account.Element, tas []tadb.Element) *handler {
	return New(
		server.NewStopper(),
		20,
		account.NewMemoryDb(acnts),
		tadb.NewMemoryDb(tas),
		sector.NewMemoryDb(),
		pairwise.NewMemoryDb(),
		token.NewMemoryDb(),
		rand.New(time.Minute),
		true,
	).(*handler)
}

// GET と POST でのアカウント情報リクエストに対応するか。
func TestNormal(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	for _, meth := range []string{"GET", "POST"} {
		acnt := newTestAccount()
		hndl := newTestHandler([]account.Element{acnt}, []tadb.Element{test_ta})

		now := time.Now()
		tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), test_ta.Id())
		hndl.tokDb.Save(tok, now.Add(time.Minute))

		r, err := http.NewRequest(meth, "https://idp.example.org/userinfo", nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Authorization", "Bearer "+tok.Id())

		w := httptest.NewRecorder()
		hndl.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Error(w.Code)
			t.Fatal(http.StatusOK)
		} else if contType, contType2 := "application/json", w.HeaderMap.Get("Content-Type"); contType2 != contType {
			t.Error(contType2)
			t.Fatal(contType)
		}

		var buff struct{ Sub, Email string }
		if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
			t.Fatal(err)
		} else if buff.Sub != acnt.Id() {
			t.Fatal(buff.Sub)
		} else if buff.Email != test_email {
			t.Error(buff.Email)
			t.Fatal(test_email)
		}
	}
}

// TA 固有アカウント ID に対応していることの検査。
func TestPairwise(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	ta := tadb.New("https://ta.example.org", nil, nil, nil, true, "")
	hndl := newTestHandler([]account.Element{acnt}, []tadb.Element{ta})

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid"), strsetutil.New("email"), ta.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := http.NewRequest("GET", "https://idp.example.org/userinfo", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer "+tok.Id())

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	} else if contType, contType2 := "application/json", w.HeaderMap.Get("Content-Type"); contType2 != contType {
		t.Error(contType2)
		t.Fatal(contType)
	}

	var buff struct{ Sub, Email string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Sub == "" || buff.Sub == acnt.Id() {
		t.Error("not pairwise")
		t.Fatal(buff.Sub)
	} else if buff.Email != test_email {
		t.Error(buff.Email)
		t.Fatal(test_email)
	}
}

// スコープ属性の展開はしないことの検査。
func TestNotUseScopeAttribute(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole(logRoot, level.ALL)
	// defer logutil.SetupConsole(logRoot, level.OFF)
	// ////////////////////////////////

	acnt := newTestAccount()
	hndl := newTestHandler([]account.Element{acnt}, []tadb.Element{test_ta})

	now := time.Now()
	tok := token.New(test_tokId, now.Add(time.Minute), acnt.Id(), strsetutil.New("openid", "email"), nil, test_ta.Id())
	hndl.tokDb.Save(tok, now.Add(time.Minute))

	r, err := http.NewRequest("GET", "https://idp.example.org/userinfo", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer "+tok.Id())

	w := httptest.NewRecorder()
	hndl.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	} else if contType, contType2 := "application/json", w.HeaderMap.Get("Content-Type"); contType2 != contType {
		t.Error(contType2)
		t.Fatal(contType)
	}

	var buff struct{ Sub, Email string }
	if err := json.NewDecoder(w.Body).Decode(&buff); err != nil {
		t.Fatal(err)
	} else if buff.Sub != acnt.Id() {
		t.Error(buff.Sub)
		t.Fatal(acnt.Id())
	} else if buff.Email != "" {
		t.Error("got scope attribute")
		t.Fatal(buff.Email)
	}
}
