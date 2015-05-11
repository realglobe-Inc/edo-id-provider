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
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"reflect"
	"testing"
)

func TestSatisfiable(t *testing.T) {
	var reqClm session.Claim
	if err := json.Unmarshal([]byte(`{
    "id_token": {
        "email": {
            "essential": true
        }
    },
    "userinfo": {
        "pds": {
            "essential": true
        }
    }
}`), &reqClm); err != nil {
		t.Fatal(err)
	}

	if ok, scop, tokAttrs, acntAttrs := satisfiable(newConsentInfo(nil, map[string]bool{"email": true, "pds": true}), nil, &reqClm); !ok {
		t.Fatal("cannot detect satisfiable")
	} else if len(scop) > 0 {
		t.Fatal(scop)
	} else if !reflect.DeepEqual(tokAttrs, map[string]bool{"email": true}) {
		t.Fatal(tokAttrs)
	} else if !reflect.DeepEqual(acntAttrs, map[string]bool{"pds": true}) {
		t.Fatal(acntAttrs)
	} else if ok, scop, tokAttrs, acntAttrs := satisfiable(newConsentInfo(map[string]bool{"email": true}, map[string]bool{"pds": true}), nil, &reqClm); !ok {
		t.Fatal("cannot detect satisfiable")
	} else if len(scop) > 0 {
		t.Fatal(scop)
	} else if !reflect.DeepEqual(tokAttrs, map[string]bool{"email": true}) {
		t.Fatal(tokAttrs)
	} else if !reflect.DeepEqual(acntAttrs, map[string]bool{"pds": true}) {
		t.Fatal(acntAttrs)
	} else if ok, scop, tokAttrs, acntAttrs := satisfiable(newConsentInfo(map[string]bool{"email": true}, map[string]bool{"pds": true}), map[string]bool{"email": true}, &reqClm); !ok {
		t.Fatal("cannot detect satisfiable")
	} else if !reflect.DeepEqual(scop, map[string]bool{"email": true}) {
		t.Fatal(scop)
	} else if !reflect.DeepEqual(tokAttrs, map[string]bool{"email": true}) {
		t.Fatal(tokAttrs)
	} else if !reflect.DeepEqual(acntAttrs, map[string]bool{"email": true, "email_verified": true, "pds": true}) {
		t.Fatal(acntAttrs)
	} else if ok, scop, tokAttrs, acntAttrs := satisfiable(newConsentInfo(nil, nil), map[string]bool{"email": true}, &reqClm); ok {
		t.Error(scop)
		t.Error(tokAttrs)
		t.Fatal(acntAttrs)
	} else if ok, scop, tokAttrs, acntAttrs := satisfiable(newConsentInfo(nil, nil), map[string]bool{"email": true}, &reqClm); ok {
		t.Error(scop)
		t.Error(tokAttrs)
		t.Fatal(acntAttrs)
	}
}
