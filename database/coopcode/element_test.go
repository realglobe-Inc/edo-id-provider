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
	test_id      = "1SblzkyNc6O867zqdZYPM0T-a7g1n5"
	test_ta_from = "https://from.example.org"
	test_ta_to   = "https://to.example.org"

	test_tok = "TM4CmjXyWQeqtasbRDqwSN80n26vuV"
)

var (
	test_acnt      = NewAccount(test_acnt_id, test_acnt_tag)
	test_scop      = map[string]bool{"openid": true, "email": true}
	test_rel_acnts = []*Account{NewAccount("76q83UV-ENtUESsw", "enemy")}
)

func TestElement(t *testing.T) {
	exp := time.Now().Add(time.Second)
	tok_exp := exp.Add(time.Hour)
	elem := New(test_id, exp, test_acnt, test_scop, tok_exp, test_rel_acnts, test_ta_from, test_ta_to)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Error(test_id)
		return
	} else if !elem.ExpiresIn().Equal(exp) {
		t.Error(elem.ExpiresIn())
		t.Error(exp)
		return
	} else if !reflect.DeepEqual(elem.Account(), test_acnt) {
		t.Error(elem.Account())
		t.Error(test_acnt)
		return
	} else if !reflect.DeepEqual(elem.Scope(), test_scop) {
		t.Error(elem.Scope())
		t.Error(test_scop)
		return
	} else if !elem.TokenExpiresIn().Equal(tok_exp) {
		t.Error(elem.TokenExpiresIn())
		t.Error(tok_exp)
		return
	} else if !reflect.DeepEqual(elem.RelatedAccounts(), test_rel_acnts) {
		t.Error(elem.RelatedAccounts())
		t.Error(test_rel_acnts)
		return
	} else if elem.TaFrom() != test_ta_from {
		t.Error(elem.TaFrom())
		t.Error(test_ta_from)
		return
	} else if elem.TaTo() != test_ta_to {
		t.Error(elem.TaTo())
		t.Error(test_ta_to)
		return
	} else if elem.Token() != "" {
		t.Error(elem.Token())
		return
	}

	date := elem.Date()
	elem.SetToken(test_tok)
	if elem.Token() != test_tok {
		t.Error(elem.Token())
		t.Error(test_tok)
	} else if elem.Date().Before(date) {
		t.Error(elem.Date())
		t.Error(date)
	}
}
