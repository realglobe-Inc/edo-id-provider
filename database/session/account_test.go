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
	"testing"
	"time"
)

func TestAccount(t *testing.T) {
	acnt := NewAccount(test_acntId, test_acntName)

	if acnt.Id() != test_acntId {
		t.Error(acnt.Id())
		t.Fatal(test_acntId)
	} else if acnt.Name() != test_acntName {
		t.Error(acnt.Name())
		t.Fatal(test_acntName)
	} else if acnt.LoggedIn() {
		t.Fatal("new account logged in")
	}

	bef := time.Now()
	acnt.Login()
	aft := time.Now()
	if !acnt.LoggedIn() {
		t.Fatal("not logged in")
	} else if bef.After(acnt.LoginDate()) {
		t.Error(acnt.LoginDate())
		t.Fatal(bef)
	} else if aft.Before(acnt.LoginDate()) {
		t.Error(acnt.LoginDate())
		t.Fatal(aft)
	}

	acnt2 := acnt.New()
	if acnt2.Id() != acnt.Id() {
		t.Error(acnt2.Id())
		t.Fatal(acnt.Id())
	} else if acnt2.Name() != acnt.Name() {
		t.Error(acnt2.Name())
		t.Fatal(acnt.Name())
	} else if acnt2.LoggedIn() {
		t.Fatal("copy account logged in")
	}
}
