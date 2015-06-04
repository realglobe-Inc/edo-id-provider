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
	tic := "-TRO_YRa1B"
	name := "edo-id-provider-tester"
	passwd := "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
	lang := "ja-JP"

	r, err := http.NewRequest("POST", "https://idp.example.org/auth/login", strings.NewReader(""+
		"ticket="+url.QueryEscape(tic)+
		"&username="+url.QueryEscape(name)+
		"&pass_type="+url.QueryEscape("STR43")+
		"&password="+url.QueryEscape(passwd)+
		"&locale="+url.QueryEscape(lang)+
		""))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if req, err := parseLoginRequest(r); err != nil {
		t.Fatal(err)
	} else if req.ticket() != tic {
		t.Error(req.ticket())
		t.Fatal(tic)
	} else if req.accountName() != name {
		t.Error(req.accountName())
		t.Fatal(name)
	} else if req.passType() != "STR43" {
		t.Error(req.passType())
		t.Fatal("STR43")
	} else if req.passInfo().password() != passwd {
		t.Error(req.passInfo().password())
		t.Fatal(passwd)
	} else if req.language() != lang {
		t.Error(req.language())
		t.Fatal(lang)
	}
}
