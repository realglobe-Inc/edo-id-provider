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

package coopcode

import (
	"reflect"
	"testing"
	"time"
)

const (
	test_id     = "1SblzkyNc6O867zqdZYPM0T-a7g1n5"
	test_srcTok = "qLCaRl3jF9fNaI7VJEXK3Tj80Kojqx"
	test_taFr   = "https://from.example.org"
	test_taTo   = "https://to.example.org"

	test_tok = "TM4CmjXyWQeqtasbRDqwSN80n26vuV"
)

var (
	test_acnt     = NewAccount(test_acntId, test_acntTag)
	test_scop     = map[string]bool{"openid": true, "email": true}
	test_relAcnts = []*Account{NewAccount("76q83UV-ENtUESsw", "enemy")}
)

func TestElement(t *testing.T) {
	exp := time.Now().Add(time.Minute)
	tokExp := time.Now().Add(time.Minute)
	elem := New(test_id, exp, test_acnt, test_srcTok, test_scop, tokExp, test_relAcnts, test_taFr, test_taTo)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	} else if !reflect.DeepEqual(elem.Account(), test_acnt) {
		t.Error(elem.Account())
		t.Fatal(test_acnt)
	} else if elem.SourceToken() != test_srcTok {
		t.Error(elem.SourceToken)
		t.Fatal(test_srcTok)
	} else if !reflect.DeepEqual(elem.Scope(), test_scop) {
		t.Error(elem.Scope())
		t.Fatal(test_scop)
	} else if !elem.TokenExpires().Equal(tokExp) {
		t.Error(elem.TokenExpires())
		t.Fatal(tokExp)
	} else if !reflect.DeepEqual(elem.Accounts(), test_relAcnts) {
		t.Error(elem.Accounts())
		t.Fatal(test_relAcnts)
	} else if elem.FromTa() != test_taFr {
		t.Error(elem.FromTa())
		t.Fatal(test_taFr)
	} else if elem.ToTa() != test_taTo {
		t.Error(elem.ToTa())
		t.Fatal(test_taTo)
	} else if elem.Token() != "" {
		t.Fatal(elem.Token())
	}

	date := elem.Date()
	elem.SetToken(test_tok)
	if elem.Token() != test_tok {
		t.Error(elem.Token())
		t.Fatal(test_tok)
	} else if elem.Date().Before(date) {
		t.Error(elem.Date())
		t.Fatal(date)
	}
}
