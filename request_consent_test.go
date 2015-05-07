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
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestConsentRequest(t *testing.T) {
	tic := "-TRO_YRa1B"
	lang := "ja-JP"

	r, err := http.NewRequest("POST", "https://idp.example.org/auth", strings.NewReader(""+
		"ticket="+url.QueryEscape(tic)+
		"&allowed_scope="+url.QueryEscape("openid email")+
		"&allowed_claims="+url.QueryEscape("pds")+
		"&denied_scope="+url.QueryEscape("offline_access")+
		"&denied_claims="+url.QueryEscape("address birthdate")+
		"&locale="+url.QueryEscape(lang)+
		""))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if req := newConsentRequest(r); req.ticket() != tic {
		t.Error(req.ticket())
		t.Fatal(tic)
	} else if ans := map[string]bool{"openid": true, "email": true}; !reflect.DeepEqual(req.allowedScope(), ans) {
		t.Error(req.allowedScope())
		t.Fatal(ans)
	} else if ans := map[string]bool{"pds": true}; !reflect.DeepEqual(req.allowedAttributes(), ans) {
		t.Error(req.allowedAttributes())
		t.Fatal(ans)
	} else if ans := map[string]bool{"offline_access": true}; !reflect.DeepEqual(req.deniedScope(), ans) {
		t.Error(req.deniedScope())
		t.Fatal(ans)
	} else if ans := map[string]bool{"address": true, "birthdate": true}; !reflect.DeepEqual(req.deniedAttributes(), ans) {
		t.Error(req.deniedAttributes())
		t.Fatal(ans)
	} else if req.language() != lang {
		t.Error(req.language())
		t.Fatal(lang)
	}
}
