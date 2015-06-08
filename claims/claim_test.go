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

func TestClaimSample(t *testing.T) {
	var ent Claim

	if err := json.Unmarshal([]byte(`null`), &ent); err != nil {
		t.Fatal(err)
	} else if ent.Essential() {
		t.Fatal(ent.Essential())
	} else if ent.Value() != nil {
		t.Fatal(ent.Value())
	} else if ent.Values() != nil {
		t.Fatal(ent.Values())
	}

	if err := json.Unmarshal([]byte(`{"essential": true}`), &ent); err != nil {
		t.Fatal(err)
	} else if !ent.Essential() {
		t.Fatal(ent)
	} else if ent.Value() != nil {
		t.Fatal(ent.Value())
	} else if ent.Values() != nil {
		t.Fatal(ent.Values())
	}

	if err := json.Unmarshal([]byte(`{"values": ["urn:mace:incommon:iap:silver"] }`), &ent); err != nil {
		t.Fatal(err)
	} else if ent.Essential() {
		t.Fatal(ent.Essential())
	} else if ent.Value() != nil {
		t.Fatal(ent.Value())
	} else if !reflect.DeepEqual(ent.Values(), []interface{}{"urn:mace:incommon:iap:silver"}) {
		t.Fatal(ent.Values())
	}

}
