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

package token

import (
	"reflect"
	"testing"
	"time"
)

const (
	test_id   = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acnt = "EYClXo4mQKwSgPel"
	test_ta   = "https://ta.example.org"
	test_tok  = "TM4CmjXyWQeqtasbRDqwSN80n26vuV"
)

var (
	test_scop  = map[string]bool{"openid": true, "email": true}
	test_attrs = map[string]bool{"pds": true}
)

func TestElement(t *testing.T) {
	exp := time.Now().Add(time.Second)
	elem := New(test_id, exp, test_acnt, test_scop, test_attrs, test_ta)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if elem.Invalid() {
		t.Fatal("invalid")
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	} else if elem.Account() != test_acnt {
		t.Error(elem.Account())
		t.Fatal(test_acnt)
	} else if !reflect.DeepEqual(elem.Scope(), test_scop) {
		t.Error(elem.Scope())
		t.Fatal(test_scop)
	} else if !reflect.DeepEqual(elem.Attributes(), test_attrs) {
		t.Error(elem.Attributes())
		t.Fatal(test_attrs)
	} else if elem.Ta() != test_ta {
		t.Error(elem.Ta())
		t.Fatal(test_ta)
	} else if len(elem.Tokens()) > 0 {
		t.Fatal(elem.Tokens())
	}

	date := elem.Date()
	elem.AddToken(test_tok)
	if !reflect.DeepEqual(elem.Tokens(), map[string]bool{test_tok: true}) {
		t.Error(elem.Tokens())
		t.Fatal(map[string]bool{test_tok: true})
	} else if elem.Date().Before(date) {
		t.Error(elem.Date())
		t.Fatal(date)
	}

	date = elem.Date()
	elem.Invalidate()
	if !elem.Invalid() {
		t.Fatal("valid after invalidate")
	} else if elem.Date().Before(date) {
		t.Error(elem.Date())
		t.Fatal(date)
	}
}
