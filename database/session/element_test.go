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
	test_acnt = NewAccount(test_acnt_id, test_acnt_name)
	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		panic(err)
	}
	q := url.Values{}
	q.Add("scope", "openid email")
	q.Add("response_type", "code id_token")
	q.Add("client_id", test_ta)
	q.Add("redirect_uri", test_redi_uri.String())
	q.Add("state", test_stat)
	q.Add("nonce", test_nonc)
	q.Add("display", test_disp)
	q.Add("prompt", "login consent")
	q.Add("max_age", strconv.FormatInt(int64(test_max_age/time.Second), 10))
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
		t.Error(test_id)
		return
	} else if !elem.ExpiresIn().Equal(exp) {
		t.Error(elem.ExpiresIn())
		t.Error(exp)
		return
	} else if elem.Request() != nil {
		t.Error(elem.Request())
		return
	} else if elem.Ticket() != "" {
		t.Error(elem.Ticket())
		return
	} else if len(elem.SelectedAccounts()) > 0 {
		t.Error(elem.SelectedAccounts())
		return
	} else if elem.Language() != "" {
		t.Error(elem.Language())
		return
	}

	elem.SelectAccount(test_acnt)
	elem.SetRequest(test_req)
	elem.SetTicket(test_tic)
	elem.SetLanguage(test_lang)

	if !reflect.DeepEqual(elem.Request(), test_req) {
		t.Error(elem.Request())
		t.Error(test_req)
		return
	} else if elem.Ticket() != test_tic {
		t.Error(elem.Ticket())
		t.Error(test_tic)
		return
	} else if !reflect.DeepEqual(elem.SelectedAccounts(), []*Account{test_acnt}) {
		t.Error(elem.SelectedAccounts())
		t.Error([]*Account{test_acnt})
		return
	} else if elem.Language() != test_lang {
		t.Error(elem.Language())
		t.Error(test_lang)
		return
	}

	elem.Clear()
	if elem.Request() != nil {
		t.Error(elem.Request())
	} else if elem.Ticket() != "" {
		t.Error(elem.Ticket())
	}
}

func TestElementPastAccount(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	elem := New(test_id, exp)
	if len(elem.SelectedAccounts()) != 0 {
		t.Error(elem.SelectedAccounts())
		return
	}

	elem.SelectAccount(NewAccount(test_acnt_id, test_acnt_name))
	if len(elem.SelectedAccounts()) != 1 {
		t.Error(elem.SelectedAccounts())
		return
	}

	elem.SelectAccount(NewAccount(test_acnt_id+"2", test_acnt_name+"2"))
	if len(elem.SelectedAccounts()) != 2 {
		t.Error(elem.SelectedAccounts())
		return
	}

	// 同じのだから増えない。
	elem.SelectAccount(NewAccount(test_acnt_id, test_acnt_name))
	if len(elem.SelectedAccounts()) != 2 {
		t.Error(elem.SelectedAccounts())
		return
	}

	elem.SelectAccount(NewAccount(test_acnt_id+"3", test_acnt_name+"3"))
	if len(elem.SelectedAccounts()) != 3 {
		t.Error(elem.SelectedAccounts())
		return
	}

	if !reflect.DeepEqual(elem.SelectedAccounts(), []*Account{
		NewAccount(test_acnt_id+"3", test_acnt_name+"3"),
		NewAccount(test_acnt_id, test_acnt_name),
		NewAccount(test_acnt_id+"2", test_acnt_name+"2")}) {
		t.Error(elem.SelectedAccounts())
		return
	}

	for i := 0; i < 2*MaxHistory; i++ {
		elem.SelectAccount(NewAccount(test_acnt_id+strconv.Itoa(i), test_acnt_name+strconv.Itoa(i)))
		if acnts := elem.SelectedAccounts(); len(acnts) > MaxHistory+1 {
			t.Error(i)
			t.Error(acnts)
			return
		}
		elem.Account().Login()
	}
	if acnts := elem.SelectedAccounts(); len(acnts) != MaxHistory {
		t.Error(acnts)
		return
	}
	for _, acnt := range elem.SelectedAccounts() {
		if !acnt.LoggedIn() {
			t.Error(acnt)
			return
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
		elem.SelectAccount(NewAccount(test_acnt_id+strconv.Itoa(i), test_acnt_name+strconv.Itoa(i)))
		elem2 := elem.New(test_id+"2", exp.Add(time.Second))

		if elem2.Id() == elem.Id() {
			t.Error(i)
			t.Error(elem2.Id())
			return
		} else if elem2.ExpiresIn().Equal(elem.ExpiresIn()) {
			t.Error(i)
			t.Error(elem2.ExpiresIn())
			return
		} else if elem2.Account() != nil {
			t.Error(i)
			t.Error(elem2.Account())
			return
		} else if elem2.Request() != nil {
			t.Error(i)
			t.Error(elem2.Request())
			return
		} else if elem2.Ticket() != "" {
			t.Error(i)
			t.Error(elem2.Ticket())
			return
		} else if !reflect.DeepEqual(elem.SelectedAccounts(), elem2.SelectedAccounts()) {
			t.Error(i)
			t.Error(elem2.SelectedAccounts())
			t.Error(elem.SelectedAccounts())
			return
		} else if elem2.Language() != elem.Language() {
			t.Error(i)
			t.Error(elem2.Language())
			t.Error(elem.Language())
			return
		}
	}

	for i := 0; i < MaxHistory; i++ {
		elem.SelectAccount(NewAccount(test_acnt_id+strconv.Itoa(i), test_acnt_name+strconv.Itoa(i)))
		elem.Account().Login()
		elem2 := elem.New(test_id+"2", exp.Add(time.Second))

		for _, acnt := range elem2.SelectedAccounts() {
			if acnt.LoggedIn() {
				t.Error(i)
				t.Error("new account logged in")
				return
			}
		}
	}
}
