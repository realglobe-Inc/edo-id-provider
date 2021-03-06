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

package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestLoginRequest(t *testing.T) {
	q := url.Values{}
	q.Set("ticket", test_ticId)
	q.Set("username", test_acntName)
	q.Set("pass_type", "password")
	q.Set("password", test_passwd)
	q.Set("locale", test_lang)
	r, err := http.NewRequest("POST", "https://idp.example.org/auth/login", strings.NewReader(q.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if req, err := parseLoginRequest(r); err != nil {
		t.Fatal(err)
	} else if req.ticket() != test_ticId {
		t.Error(req.ticket())
		t.Fatal(test_ticId)
	} else if req.accountName() != test_acntName {
		t.Error(req.accountName())
		t.Fatal(test_acntName)
	} else if req.passInfo().passType() != "password" {
		t.Error(req.passInfo().passType())
		t.Fatal("password")
	} else if req.passInfo().params()[0] != test_passwd {
		t.Error(req.passInfo().params()[0])
		t.Fatal(test_passwd)
	} else if req.language() != test_lang {
		t.Error(req.language())
		t.Fatal(test_lang)
	}
}
