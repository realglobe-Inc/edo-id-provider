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

package consent

import (
	"testing"
)

const (
	test_acnt = "EYClXo4mQKwSgPel"
	test_ta   = "https://ta.example.org"
	test_scop = "openid"
	test_attr = "pds"
)

func TestElement(t *testing.T) {
	a := New(test_acnt, test_ta)
	if a.Account() != test_acnt {
		t.Fatal(a.Account())
	} else if a.Ta() != test_ta {
		t.Fatal(a.Ta())
	}

	if a.ScopeAllowed(test_scop) {
		t.Fatal(a)
	} else if a.AttributeAllowed(test_attr) {
		t.Fatal(a)
	}

	a.AllowScope(test_scop)
	a.AllowAttribute(test_attr)
	if !a.ScopeAllowed(test_scop) {
		t.Fatal(a)
	} else if !a.AttributeAllowed(test_attr) {
		t.Fatal(a)
	}

	a.DenyScope(test_scop)
	a.DenyAttribute(test_attr)
	if a.ScopeAllowed(test_scop) {
		t.Fatal(a)
	} else if a.AttributeAllowed(test_attr) {
		t.Fatal(a)
	}
}
