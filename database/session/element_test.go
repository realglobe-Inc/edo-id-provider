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
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestElementPastAccount(t *testing.T) {
	a := New("test-session-id", time.Date(2015, time.April, 4, 18, 41, 20, 123456789, time.UTC))
	if acnts := a.SelectedAccounts(); len(acnts) != 0 {
		t.Error(acnts)
		return
	}

	a.SelectAccount(NewAccount("test-account", "test-account-name"))
	if acnts := a.SelectedAccounts(); len(acnts) != 1 {
		t.Error(acnts)
		return
	}

	a.SelectAccount(NewAccount("test-account2", "test-account-name2"))
	if acnts := a.SelectedAccounts(); len(acnts) != 2 {
		t.Error(acnts)
		return
	}

	a.SelectAccount(NewAccount("test-account", "test-account-name"))
	if acnts := a.SelectedAccounts(); len(acnts) != 2 {
		t.Error(acnts)
		return
	}

	a.SelectAccount(NewAccount("test-account3", "test-account-name3"))
	if acnts := a.SelectedAccounts(); len(acnts) != 3 {
		t.Error(acnts)
		return
	}

	if acnts := a.SelectedAccounts(); !reflect.DeepEqual(acnts, []*Account{NewAccount("test-account3", "test-account-name3"), NewAccount("test-account", "test-account-name"), NewAccount("test-account2", "test-account-name2")}) {
		t.Error(acnts)
		return
	}

	for i := 0; i < 2*MaxHistory; i++ {
		a.SelectAccount(NewAccount("test-account"+strconv.Itoa(i), "test-account-name"+strconv.Itoa(i)))
		if acnts := a.SelectedAccounts(); len(acnts) > MaxHistory+1 {
			t.Error(i)
			t.Error(acnts)
			return
		}
		a.Account().Login()
	}
	if acnts := a.SelectedAccounts(); len(acnts) != MaxHistory {
		t.Error(acnts)
		return
	}
	for _, acnt := range a.SelectedAccounts() {
		if !acnt.LoggedIn() {
			t.Error(acnt)
			return
		}
	}
}

func TestElementNew(t *testing.T) {
	// OpenID Connect Core 1.0 Section 3.1.2.1 より。
	rawReq, err := http.NewRequest("GET", "https://server.example.com/authorize?"+
		"response_type=code"+
		"&scope=openid%20profile%20email"+
		"&client_id=s6BhdRkqt3"+
		"&state=af0ifjsldkj"+
		"&redirect_uri=https%3A%2F%2Fclient.example.org%2Fcb", nil)
	if err != nil {
		t.Error(err)
		return
	}
	req, err := ParseRequest(rawReq)
	if err != nil {
		t.Error(err)
		return
	}

	date := time.Date(2015, time.April, 4, 18, 41, 20, 123456789, time.UTC)
	a := New("test-session-id", date)
	a.SetRequest(req)
	a.SetTicket("test-ticket")
	a.SetLanguage("test-language")
	for i := 0; i < 2*MaxHistory; i++ {
		a.SelectAccount(NewAccount("test-account"+strconv.Itoa(i), "test-account-name"+strconv.Itoa(i)))
		b := a.New("test-session-id2", date.Add(time.Second))

		if b.Id() == a.Id() {
			t.Error(i)
			t.Error(b.Id())
			return
		} else if b.ExpiresIn().Equal(a.ExpiresIn()) {
			t.Error(i)
			t.Error(b.ExpiresIn())
			return
		} else if b.Account() != nil {
			t.Error(i)
			t.Error(b.Account())
			return
		} else if b.Request() != nil {
			t.Error(i)
			t.Error(b.Request())
			return
		} else if b.Ticket() != "" {
			t.Error(i)
			t.Error(b.Ticket())
			return
		} else if acnts, acnts2 := a.SelectedAccounts(), b.SelectedAccounts(); !reflect.DeepEqual(acnts, acnts2) {
			t.Error(i)
			t.Error(acnts2)
			t.Error(acnts)
			return
		} else if b.Language() != a.Language() {
			t.Error(i)
			t.Error(b.Language())
			t.Error(a.Language())
			return
		}
	}

	for i := 0; i < MaxHistory; i++ {
		a.SelectAccount(NewAccount("test-account"+strconv.Itoa(i), "test-account-name"+strconv.Itoa(i)))
		a.Account().Login()
		b := a.New("test-session-id2", date.Add(time.Second))

		for _, acnt := range b.SelectedAccounts() {
			if acnt.LoggedIn() {
				t.Error(i)
				t.Error("new account logged in")
				return
			}
		}
	}
}
