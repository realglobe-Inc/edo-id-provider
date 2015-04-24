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

const (
	test_acnt_id   = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acnt_name = "edo-id-provider-tester"
)

func TestAccount(t *testing.T) {
	acnt := NewAccount(test_acnt_id, test_acnt_name)

	if acnt.Id() != test_acnt_id {
		t.Error(acnt.Id())
		t.Error(test_acnt_id)
		return
	} else if acnt.Name() != test_acnt_name {
		t.Error(acnt.Name())
		t.Error(test_acnt_name)
		return
	} else if acnt.LoggedIn() {
		t.Error("new account logged in")
		return
	}

	bef := time.Now()
	acnt.Login()
	aft := time.Now()
	if !acnt.LoggedIn() {
		t.Error("not logged in")
		return
	} else if bef.After(acnt.LoginDate()) {
		t.Error(acnt.LoginDate())
		t.Error(bef)
		return
	} else if aft.Before(acnt.LoginDate()) {
		t.Error(acnt.LoginDate())
		t.Error(aft)
		return
	}

	acnt2 := acnt.New()
	if acnt2.Id() != acnt.Id() {
		t.Error(acnt2.Id())
		t.Error(acnt.Id())
	} else if acnt2.Name() != acnt.Name() {
		t.Error(acnt2.Name())
		t.Error(acnt.Name())
	} else if acnt2.LoggedIn() {
		t.Error("copy account logged in")
	}
}
