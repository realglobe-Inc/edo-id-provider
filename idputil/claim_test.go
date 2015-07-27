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

package idputil

import (
	"encoding/json"
	"testing"

	"github.com/realglobe-Inc/edo-id-provider/claims"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
)

func TestCheckClaims(t *testing.T) {
	var reqClm claims.Claims
	if err := json.Unmarshal([]byte(`{
    "email": {
        "value": "tester@example.org"
    },
    "pds": {
        "essential": true
    }
}`), &reqClm); err != nil {
		t.Fatal(err)
	}

	if err := CheckClaims(account.New("EYClXo4mQKwSgPel", "edo-id-provider-tester", nil, map[string]interface{}{
		"email": "tester@example.org",
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}), reqClm); err != nil {
		t.Fatal(err)
	} else if err := CheckClaims(account.New("EYClXo4mQKwSgPel", "edo-id-provider-tester", nil, map[string]interface{}{
		"email": "tester2@example.org",
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}), reqClm); err == nil {
		t.Fatal("cannot detect email.value contradiction")
	} else if err := CheckClaims(account.New("EYClXo4mQKwSgPel", "edo-id-provider-tester", nil, map[string]interface{}{
		"email": "tester@example.org",
	}), reqClm); err == nil {
		t.Fatal("cannot detect lack of essential")
	}
}
