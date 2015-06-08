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

package coopto

import (
	"bytes"
	"github.com/realglobe-Inc/edo-id-provider/claims"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRequestSample(t *testing.T) {
	r, err := http.NewRequest("POST", "https://idp.example.org/coop/to", strings.NewReader(`{
    "grant_type": "cooperation_code",
    "code": "p9-FpxXxXBt5PQrM9-6T-t3l9eSz1n",
    "claims": {
        "id_token": {
            "pds": {
                "essential": true
            }
        }
    },
    "user_claims": {
        "invitee": {
            "pds": {
                "essential": true
            }
        }
    },
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
	} else if grntType := "cooperation_code"; req.grantType() != grntType {
		t.Error(req.grantType())
		t.Fatal(grntType)
	} else if cod := "p9-FpxXxXBt5PQrM9-6T-t3l9eSz1n"; req.code() != cod {
		t.Error(req.code())
		t.Fatal(cod)
	} else if clmReq := claims.NewRequest(claims.Claims{"pds": claims.New(true, nil, nil, "")}, nil); !reflect.DeepEqual(req.claims(), clmReq) {
		t.Error(req.claims())
		t.Fatal(clmReq)
	} else if subClmReqs := map[string]claims.Claims{"invitee": {"pds": claims.New(true, nil, nil, "")}}; !reflect.DeepEqual(req.subClaims(), subClmReqs) {
		t.Error(req.subClaims())
		t.Fatal(subClmReqs)
	} else if assType := "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"; req.taAssertionType() != assType {
		t.Error(req.taAssertionType())
		t.Fatal(assType)
	} else if ass := []byte("eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZy90b2tlbiIsImV4cCI6MTQyNTQ1MzI2MywiaWF0IjoxNDI1NDUyNjYzLCJpc3MiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIiwianRpIjoiNkwzZ3RPTF9jc3pyX1RKSWVkWlciLCJzdWIiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIn0.Q3UA2dpNLgrvk4SL5-9zXH_aSA2_OQX7ixfgnxKbGoF0W0YpyRHwYQu1N-4MXKHAQbaVo-1cH7UIAv1PcT6-2A"); !bytes.Equal(req.taAssertion(), ass) {
		t.Error(string(req.taAssertion()))
		t.Fatal(string(ass))
	}
}
