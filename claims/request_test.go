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

package claims

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
)

func TestRequestSample(t *testing.T) {
	// OpenID Connect Core 1.0 Section 5.5 より。
	sample := []byte(`
{
   "userinfo":
    {
     "given_name": {"essential": true},
     "nickname": null,
     "email": {"essential": true},
     "email_verified": {"essential": true},
     "picture": null,
     "http://example.info/claims/groups": null
    },
   "id_token":
    {
     "auth_time": {"essential": true},
     "acr": {"values": ["urn:mace:incommon:iap:silver"] }
    }
  }
`)
	var req Request
	if err := json.Unmarshal(sample, &req); err != nil {
		t.Fatal(err)
	} else if acntInf := req.AccountEntries(); acntInf == nil {
		t.Fatal(req)
	} else if v := acntInf["given_name"]; v == nil || !v.Essential() {
		t.Fatal(v)
	} else if v, ok := acntInf["nickname"]; v == nil || v.Essential() {
		t.Fatal(v, ok)
	} else if v := acntInf["email"]; v == nil || !v.Essential() {
		t.Fatal(v)
	} else if v := acntInf["email_verified"]; v == nil || !v.Essential() {
		t.Fatal(v)
	} else if v, ok := acntInf["picture"]; v == nil || v.Essential() {
		t.Fatal(v, ok)
	} else if v, ok := acntInf["http://example.info/claims/groups"]; v == nil || v.Essential() {
		t.Fatal(v, ok)
	} else if idTok := req.IdTokenEntries(); idTok == nil {
		t.Fatal(req)
	} else if v := idTok["auth_time"]; v == nil || !v.Essential() {
		t.Fatal(v)
	} else if v := idTok["acr"]; v == nil ||
		!reflect.DeepEqual(toStrings(v.Values()), []string{"urn:mace:incommon:iap:silver"}) {
		t.Fatal(v)
	}
	clms, optClms := req.Names()
	if clmSet := strsetutil.New("given_name", "email", "email_verified", "auth_time"); !reflect.DeepEqual(clms, clmSet) {
		t.Error(clms)
		t.Fatal(clmSet)
	} else if optClmSet := strsetutil.New("nickname", "picture", "http://example.info/claims/groups", "acr"); !reflect.DeepEqual(optClms, optClmSet) {
		t.Error(optClms)
		t.Fatal(optClmSet)
	}
}

func toStrings(a []interface{}) []string {
	b := []string{}
	for _, v := range a {
		b = append(b, v.(string))
	}
	return b
}

func TestRequestJson(t *testing.T) {
	req := NewRequest(Claims{"pds": New(true, nil, nil, "")}, Claims{"nickname": New(false, nil, nil, "ja-JP")})

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var req2 Request
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(&req2, req) {
		t.Error(fmt.Sprintf("%#v", &req2))
		t.Fatal(fmt.Sprintf("%#v", req))
	}
}
