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

package authcode

import (
	"reflect"
	"testing"
	"time"
)

const (
	test_id      = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acnt    = "EYClXo4mQKwSgPel"
	test_ta      = "https://ta.example.org"
	test_rediUri = "https://ta.example.org/callback"
	test_nonc    = "Wjj1_YUOlR"
	test_tok     = "TM4CmjXyWQeqtasbRDqwSN80n26vuV"
)

var (
	test_scop      = map[string]bool{"openid": true, "email": true}
	test_acntAttrs = map[string]bool{"pds": true}
)

func TestElement(t *testing.T) {
	lgin := time.Now()
	exp := time.Now().Add(time.Second)
	elem := New(test_id, exp, test_acnt, lgin, test_scop, nil, test_acntAttrs, test_ta, test_rediUri, test_nonc)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	} else if elem.Account() != test_acnt {
		t.Error(elem.Account())
		t.Fatal(test_acnt)
	} else if !reflect.DeepEqual(elem.Scope(), test_scop) {
		t.Error(elem.Scope())
		t.Fatal(test_scop)
	} else if !elem.LoginDate().Equal(lgin) {
		t.Error(elem.LoginDate())
		t.Fatal(lgin)
	} else if elem.IdTokenAttributes() != nil {
		t.Fatal(elem.IdTokenAttributes())
	} else if !reflect.DeepEqual(elem.AccountAttributes(), test_acntAttrs) {
		t.Error(elem.AccountAttributes())
		t.Fatal(test_acntAttrs)
	} else if elem.Ta() != test_ta {
		t.Error(elem.Ta())
		t.Fatal(test_ta)
	} else if !reflect.DeepEqual(elem.RedirectUri(), test_rediUri) {
		t.Error(elem.RedirectUri())
		t.Fatal(test_rediUri)
	} else if elem.Nonce() != test_nonc {
		t.Error(elem.Nonce())
		t.Fatal(test_nonc)
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
