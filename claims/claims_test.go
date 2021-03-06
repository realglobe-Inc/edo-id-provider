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
	"reflect"
	"testing"
)

func TestClaims(t *testing.T) {
	var clms Claims

	if err := json.Unmarshal([]byte(`{"test-attribute":{"essential":true}}`), &clms); err != nil {
		t.Fatal(err)
	} else if clms["test-attribute"] == nil {
		t.Fatal("no claim")
	}

	if err := json.Unmarshal([]byte(`{"test-attribute":null}`), &clms); err != nil {
		t.Fatal(err)
	} else if clms["test-attribute"] == nil {
		t.Fatal("no claim")
	}

	if err := json.Unmarshal([]byte(`{"test-attribute#ja":null}`), &clms); err != nil {
		t.Fatal(err)
	} else if ent := clms["test-attribute"]; ent == nil {
		t.Fatal("no claim")
	} else if ent.Language() != "ja" {
		t.Error(ent.Language())
		t.Fatal("ja")
	}
}

func TestClaimsJson(t *testing.T) {
	clms := Claims{"pds": New(true, nil, nil, ""), "nickname": New(false, nil, nil, "ja-JP")}

	data, err := json.Marshal(clms)
	if err != nil {
		t.Fatal(err)
	}

	var clms2 Claims
	if err := json.Unmarshal(data, &clms2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(clms2, clms) {
		t.Error(clms2)
		t.Fatal(clms)
	}
}
