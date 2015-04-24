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

package session

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestClaimSample(t *testing.T) {
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
	var clms Claim
	if err := json.Unmarshal(sample, &clms); err != nil {
		t.Error(err)
		return
	} else if acntInf := clms.WithAccountInfo(); acntInf == nil {
		t.Error(clms)
	} else if v := acntInf["given_name"]; v == nil || !v.Essential() {
		t.Error(v)
	} else if v, ok := acntInf["nickname"]; !ok || v != nil {
		t.Error(v, ok)
	} else if v := acntInf["email"]; v == nil || !v.Essential() {
		t.Error(v)
	} else if v := acntInf["email_verified"]; v == nil || !v.Essential() {
		t.Error(v)
	} else if v, ok := acntInf["picture"]; !ok || v != nil {
		t.Error(v, ok)
	} else if v, ok := acntInf["http://example.info/claims/groups"]; !ok || v != nil {
		t.Error(v, ok)
	} else if idTok := clms.WithIdToken(); idTok == nil {
		t.Error(clms)
	} else if v := idTok["auth_time"]; v == nil || !v.Essential() {
		t.Error(v)
	} else if v := idTok["acr"]; v == nil ||
		!reflect.DeepEqual(toStrings(v.Values()), []string{"urn:mace:incommon:iap:silver"}) {
		t.Error(v)
	}
}

func toStrings(a []interface{}) []string {
	b := []string{}
	for _, v := range a {
		b = append(b, v.(string))
	}
	return b
}
