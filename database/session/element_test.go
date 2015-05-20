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

package session

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const (
	test_id   = "XAOiyqgngWGzZbgl6j1w6Zm3ytHeI-"
	test_tic  = "-TRO_YRa1B"
	test_lang = "ja-JP"
)

var (
	test_acnt *Account
	test_req  *Request
)

func init() {
	test_acnt = NewAccount(test_acntId, test_acntName)
	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		panic(err)
	}
	q := url.Values{}
	q.Add("scope", "openid email")
	q.Add("response_type", "code id_token")
	q.Add("client_id", test_ta)
	q.Add("redirect_uri", test_rediUri)
	q.Add("state", test_stat)
	q.Add("nonce", test_nonc)
	q.Add("display", test_disp)
	q.Add("prompt", "login consent")
	q.Add("max_age", strconv.FormatInt(int64(test_maxAge/time.Second), 10))
	q.Add("ui_locales", "ja-JP")
	//q.Add("id_token_hint", "")
	q.Add("claims", `{"id_token":{"pds":{"essential":true}}}`)
	//q.Add("request", "")
	//q.Add("request_uri", "")
	r.URL.RawQuery = q.Encode()

	test_req, err = ParseRequest(r)
	if err != nil {
		panic(err)
	}
}

func TestElement(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	elem := New(test_id, exp)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	} else if elem.Account() != nil {
		t.Fatal(elem.Account())
	} else if elem.Request() != nil {
		t.Fatal(elem.Request())
	} else if elem.Ticket() != "" {
		t.Fatal(elem.Ticket())
	} else if len(elem.SelectedAccounts()) > 0 {
		t.Fatal(elem.SelectedAccounts())
	} else if elem.Language() != "" {
		t.Fatal(elem.Language())
	}

	elem.SelectAccount(test_acnt)
	elem.SetRequest(test_req)
	elem.SetTicket(test_tic)
	elem.SetLanguage(test_lang)

	if !reflect.DeepEqual(elem.Account(), test_acnt) {
		t.Error(elem.Account())
		t.Fatal(test_acnt)
	} else if !reflect.DeepEqual(elem.Request(), test_req) {
		t.Error(elem.Request())
		t.Fatal(test_req)
	} else if elem.Ticket() != test_tic {
		t.Error(elem.Ticket())
		t.Fatal(test_tic)
	} else if !reflect.DeepEqual(elem.SelectedAccounts(), []*Account{test_acnt}) {
		t.Error(elem.SelectedAccounts())
		t.Fatal([]*Account{test_acnt})
	} else if elem.Language() != test_lang {
		t.Error(elem.Language())
		t.Fatal(test_lang)
	}

	elem.Clear()
	if elem.Request() != nil {
		t.Fatal(elem.Request())
	} else if elem.Ticket() != "" {
		t.Fatal(elem.Ticket())
	}
}

func TestElementPastAccount(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	elem := New(test_id, exp)
	if len(elem.SelectedAccounts()) != 0 {
		t.Fatal(elem.SelectedAccounts())
	}

	elem.SelectAccount(NewAccount(test_acntId, test_acntName))
	if len(elem.SelectedAccounts()) != 1 {
		t.Fatal(elem.SelectedAccounts())
	}

	elem.SelectAccount(NewAccount(test_acntId+"2", test_acntName+"2"))
	if len(elem.SelectedAccounts()) != 2 {
		t.Fatal(elem.SelectedAccounts())
	}

	// 同じのだから増えない。
	elem.SelectAccount(NewAccount(test_acntId, test_acntName))
	if len(elem.SelectedAccounts()) != 2 {
		t.Fatal(elem.SelectedAccounts())
	}

	elem.SelectAccount(NewAccount(test_acntId+"3", test_acntName+"3"))
	if len(elem.SelectedAccounts()) != 3 {
		t.Fatal(elem.SelectedAccounts())
	}

	if !reflect.DeepEqual(elem.SelectedAccounts(), []*Account{
		NewAccount(test_acntId+"3", test_acntName+"3"),
		NewAccount(test_acntId, test_acntName),
		NewAccount(test_acntId+"2", test_acntName+"2")}) {
		t.Fatal(elem.SelectedAccounts())
	}

	for i := 0; i < 2*MaxHistory; i++ {
		elem.SelectAccount(NewAccount(test_acntId+strconv.Itoa(i), test_acntName+strconv.Itoa(i)))
		if len(elem.SelectedAccounts()) > MaxHistory {
			t.Error(i)
			t.Fatal(elem.SelectedAccounts())
		}
		elem.Account().Login()
	}
	if len(elem.SelectedAccounts()) != MaxHistory {
		t.Fatal(elem.SelectedAccounts())
	}
	for _, acnt := range elem.SelectedAccounts() {
		if !acnt.LoggedIn() {
			t.Fatal(acnt)
		}
	}
}

func TestElementNew(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	elem := New(test_id, exp)
	elem.SetRequest(test_req)
	elem.SetTicket(test_tic)
	elem.SetLanguage(test_lang)
	for i := 0; i < 2*MaxHistory; i++ {
		elem.SelectAccount(NewAccount(test_acntId+strconv.Itoa(i), test_acntName+strconv.Itoa(i)))
		elem2 := elem.New(test_id+"2", exp.Add(time.Second))

		if elem2.Id() == elem.Id() {
			t.Error(i)
			t.Fatal(elem2.Id())
		} else if elem2.Expires().Equal(elem.Expires()) {
			t.Error(i)
			t.Fatal(elem2.Expires())
		} else if elem2.Account() != nil {
			t.Error(i)
			t.Fatal(elem2.Account())
		} else if elem2.Request() != nil {
			t.Error(i)
			t.Fatal(elem2.Request())
		} else if elem2.Ticket() != "" {
			t.Error(i)
			t.Fatal(elem2.Ticket())
		} else if !reflect.DeepEqual(elem.SelectedAccounts(), elem2.SelectedAccounts()) {
			t.Error(i)
			t.Error(elem2.SelectedAccounts())
			t.Fatal(elem.SelectedAccounts())
		} else if elem2.Language() != elem.Language() {
			t.Error(i)
			t.Error(elem2.Language())
			t.Fatal(elem.Language())
		}
	}

	for i := 0; i < MaxHistory; i++ {
		elem.SelectAccount(NewAccount(test_acntId+strconv.Itoa(i), test_acntName+strconv.Itoa(i)))
		elem.Account().Login()
		elem2 := elem.New(test_id+"2", exp.Add(time.Second))

		for _, acnt := range elem2.SelectedAccounts() {
			if acnt.LoggedIn() {
				t.Error(i)
				t.Fatal("new account logged in")
			}
		}
	}
}
