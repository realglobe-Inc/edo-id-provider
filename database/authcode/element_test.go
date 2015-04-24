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
	"net/url"
	"reflect"
	"testing"
	"time"
)

const (
	test_id   = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acnt = "EYClXo4mQKwSgPel"
	test_ta   = "https://ta.example.org"
	test_nonc = "Wjj1_YUOlR"
	test_tok  = "TM4CmjXyWQeqtasbRDqwSN80n26vuV"
)

var (
	test_scop     = map[string]bool{"openid": true, "email": true}
	test_attrs    = map[string]bool{"pds": true}
	test_redi_uri *url.URL
)

func init() {
	var err error
	test_redi_uri, err = url.Parse("https://ta.example.org/return")
	if err != nil {
		panic(err)
	}
}

func TestElement(t *testing.T) {
	exp := time.Now().Add(time.Second)
	elem := New(test_id, exp, test_acnt, test_scop, test_attrs, test_ta, test_redi_uri, test_nonc)

	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Error(test_id)
		return
	} else if !elem.ExpiresIn().Equal(exp) {
		t.Error(elem.ExpiresIn())
		t.Error(exp)
		return
	} else if elem.Account() != test_acnt {
		t.Error(elem.Account())
		t.Error(test_acnt)
		return
	} else if !reflect.DeepEqual(elem.Scope(), test_scop) {
		t.Error(elem.Scope())
		t.Error(test_scop)
		return
	} else if !reflect.DeepEqual(elem.Attributes(), test_attrs) {
		t.Error(elem.Attributes())
		t.Error(test_attrs)
		return
	} else if elem.Ta() != test_ta {
		t.Error(elem.Ta())
		t.Error(test_ta)
		return
	} else if !reflect.DeepEqual(elem.RedirectUri(), test_redi_uri) {
		t.Error(elem.RedirectUri())
		t.Error(test_redi_uri)
		return
	} else if elem.Nonce() != test_nonc {
		t.Error(elem.Nonce())
		t.Error(test_nonc)
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
