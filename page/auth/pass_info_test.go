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

func TestPasswordInfo(t *testing.T) {
	pass := newPasswordInfo(test_passwd)
	if len(pass.params()) != 1 {
		t.Fatal(pass.params())
	} else if pass.params()[0] != test_passwd {
		t.Error(pass.params())
		t.Fatal(test_passwd)
	}
}

func TestParsePassInfo(t *testing.T) {
	q := url.Values{}
	q.Set("pass_type", "password")
	q.Set("password", test_passwd)
	r, err := http.NewRequest("POST", "https://idp.example.org/login", strings.NewReader(q.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	pass, err := parsePassInfo(r)
	if err != nil {
		t.Fatal(err)
	}

	if pass.passType() != "password" {
		t.Error(pass.passType())
		t.Fatal("password")
	} else if len(pass.params()) != 1 {
		t.Fatal(pass.params())
	} else if pass.params()[0] != test_passwd {
		t.Error(pass.params()[0])
		t.Fatal(test_passwd)
	}
}
