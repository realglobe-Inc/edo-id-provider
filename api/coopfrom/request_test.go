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

package coopfrom

import (
	"bytes"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRequestSample(t *testing.T) {
	r, err := http.NewRequest("POST", "https://idp.example.org/coop/from", strings.NewReader(`{
    "response_type": "code_token referral",
    "from_client": "https://from.example.org",
    "to_client": "https://to.example.org",
    "grant_type": "access_token",
    "access_token": "cYcFjo0EF7FiN8Jx1NJ8Wn51gcYl84",
    "user_tag": "inviter",
    "users": {
        "invitee": "md04LUMHnwLYodm8e55hc8UbGISJc4ZCYJK3AWF5IBk"
    },
    "related_users": {
        "observer": "Gbr93kvtHPkuUX4YlRD4QA"
    },
    "hash_alg": "SHA256",
    "related_issuers": [
        "https://idp2.example.org"
    ],
    "client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
    "client_assertion": "eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZy90b2tlbiIsImV4cCI6MTQyNTQ1MzI2MywiaWF0IjoxNDI1NDUyNjYzLCJpc3MiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIiwianRpIjoiNkwzZ3RPTF9jc3pyX1RKSWVkWlciLCJzdWIiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIn0.Q3UA2dpNLgrvk4SL5-9zXH_aSA2_OQX7ixfgnxKbGoF0W0YpyRHwYQu1N-4MXKHAQbaVo-1cH7UIAv1PcT6-2A"
}`))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	req, err := parseRequest(r)
	if err != nil {
		t.Fatal(err)
	} else if grntType := "access_token"; req.grantType() != grntType {
		t.Error(req.grantType())
		t.Fatal(grntType)
	} else if respType := strsetutil.New("code_token", "referral"); !reflect.DeepEqual(req.responseType(), respType) {
		t.Error(req.responseType())
		t.Fatal(respType)
	} else if frTa := "https://from.example.org"; req.fromTa() != frTa {
		t.Error(req.fromTa())
		t.Fatal(frTa)
	} else if toTa := "https://to.example.org"; req.toTa() != toTa {
		t.Error(req.toTa())
		t.Fatal(toTa)
	} else if tok := "cYcFjo0EF7FiN8Jx1NJ8Wn51gcYl84"; req.accessToken() != tok {
		t.Error(req.accessToken())
		t.Fatal(tok)
	} else if req.scope() != nil {
		t.Fatal("scope is exist")
	} else if expIn := time.Duration(0); req.expiresIn() != expIn {
		t.Error(req.expiresIn())
		t.Fatal(expIn)
	} else if acntTag := "inviter"; req.accountTag() != acntTag {
		t.Error(req.accountTag())
		t.Fatal(acntTag)
	} else if acnts := map[string]string{"invitee": "md04LUMHnwLYodm8e55hc8UbGISJc4ZCYJK3AWF5IBk"}; !reflect.DeepEqual(req.accounts(), acnts) {
		t.Error(req.accounts())
		t.Fatal(acnts)
	} else if hashAlg := "SHA256"; req.hashAlgorithm() != hashAlg {
		t.Error(req.hashAlgorithm())
		t.Fatal(hashAlg)
	} else if relAcnts := map[string]string{"observer": "Gbr93kvtHPkuUX4YlRD4QA"}; !reflect.DeepEqual(req.relatedAccounts(), relAcnts) {
		t.Error(req.relatedAccounts())
		t.Fatal(relAcnts)
	} else if relIdps := []string{"https://idp2.example.org"}; !reflect.DeepEqual(req.relatedIdProviders(), relIdps) {
		t.Error(req.relatedIdProviders())
		t.Fatal(relIdps)
	} else if assType := "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"; req.taAssertionType() != assType {
		t.Error(req.taAssertionType())
		t.Fatal(assType)
	} else if ass := []byte("eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZy90b2tlbiIsImV4cCI6MTQyNTQ1MzI2MywiaWF0IjoxNDI1NDUyNjYzLCJpc3MiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIiwianRpIjoiNkwzZ3RPTF9jc3pyX1RKSWVkWlciLCJzdWIiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIn0.Q3UA2dpNLgrvk4SL5-9zXH_aSA2_OQX7ixfgnxKbGoF0W0YpyRHwYQu1N-4MXKHAQbaVo-1cH7UIAv1PcT6-2A"); !bytes.Equal(req.taAssertion(), ass) {
		t.Error(string(req.taAssertion()))
		t.Fatal(string(ass))
	}
}
