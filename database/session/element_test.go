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
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-idp-selector/ticket"
)

func TestElement(t *testing.T) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
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
	} else if elem.Ticket() != nil {
		t.Fatal(elem.Ticket())
	} else if len(elem.SelectedAccounts()) > 0 {
		t.Fatal(elem.SelectedAccounts())
	} else if elem.Language() != "" {
		t.Fatal(elem.Language())
	}

	tic := ticket.New(test_ticId, now.Add(time.Minute))
	elem.SelectAccount(test_acnt)
	elem.SetRequest(test_req)
	elem.SetTicket(tic)
	elem.SetLanguage(test_lang)

	if !reflect.DeepEqual(elem.Account(), test_acnt) {
		t.Error(elem.Account())
		t.Fatal(test_acnt)
	} else if !reflect.DeepEqual(elem.Request(), test_req) {
		t.Error(elem.Request())
		t.Fatal(test_req)
	} else if !reflect.DeepEqual(elem.Ticket(), tic) {
		t.Error(elem.Ticket())
		t.Fatal(tic)
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
	} else if elem.Ticket() != nil {
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
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	tic := ticket.New(test_ticId, now.Add(time.Minute))
	elem := New(test_id, exp)
	elem.SetRequest(test_req)
	elem.SetTicket(tic)
	elem.SetLanguage(test_lang)
	for i := 0; i < 2*MaxHistory; i++ {
		elem.SelectAccount(NewAccount(test_acntId+strconv.Itoa(i), test_acntName+strconv.Itoa(i)))
		elem2 := elem.New(test_id+"2", exp.Add(time.Minute))

		if elem2.Id() == elem.Id() {
			t.Error(i)
			t.Error(elem2.Id())
			t.Fatal(elem.Id())
		} else if elem2.Expires().Equal(elem.Expires()) {
			t.Error(i)
			t.Error(elem2.Expires())
			t.Fatal(elem.Expires())
		} else if elem2.Account() != nil {
			t.Error(i)
			t.Fatal(elem2.Account())
		} else if elem2.Request() != nil {
			t.Error(i)
			t.Fatal(elem2.Request())
		} else if elem2.Ticket() != nil {
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
		elem2 := elem.New(test_id+"2", exp.Add(time.Minute))

		for _, acnt := range elem2.SelectedAccounts() {
			if acnt.LoggedIn() {
				t.Error(i)
				t.Fatal("new account logged in")
			}
		}
	}
}
