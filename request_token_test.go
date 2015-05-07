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

package main

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestTokenRequest(t *testing.T) {
	grntType := "authorization_code"
	cod := "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	ta := "https://ta.example.org"
	rediUri := "https://ta.example.org/callback"
	taAssType := "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	taAss := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	r, err := http.NewRequest("POST", "https://idp.example.org/auth/login", strings.NewReader(""+
		"grant_type="+url.QueryEscape(grntType)+
		"&code="+url.QueryEscape(cod)+
		"&client_id="+url.QueryEscape(ta)+
		"&redirect_uri="+url.QueryEscape(rediUri)+
		"&client_assertion_type="+url.QueryEscape(taAssType)+
		"&client_assertion="+url.QueryEscape(taAss)+
		""))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if req := newTokenRequest(r); req.grantType() != grntType {
		t.Error(req.grantType())
		t.Fatal(grntType)
	} else if req.code() != cod {
		t.Error(req.code())
		t.Fatal(cod)
	} else if req.ta() != ta {
		t.Error(req.ta())
		t.Fatal(ta)
	} else if req.redirectUri() != rediUri {
		t.Error(req.redirectUri())
		t.Fatal(rediUri)
	} else if req.taAssertionType() != taAssType {
		t.Error(req.taAssertionType())
		t.Fatal(taAssType)
	} else if !bytes.Equal(req.taAssertion(), []byte(taAss)) {
		t.Error(req.taAssertion())
		t.Fatal(taAss)
	}
}
