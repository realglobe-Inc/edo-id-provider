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

package sector

import (
	"reflect"
	"testing"
)

func testDb(t *testing.T, db Db) {
	elem := New(test_id, test_salt)
	elem2 := New(test_id, append([]byte{0}, test_salt...))

	if el, err := db.Get(elem.Id()); err != nil {
		t.Fatal(err)
	} else if el != nil {
		t.Fatal(el)
	} else if el, err := db.SaveIfAbsent(elem); err != nil {
		t.Fatal(err)
	} else if el != nil {
		t.Fatal(el)
	} else if el, err := db.Get(elem.Id()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(el, elem) {
		t.Error(el)
		t.Fatal(elem)
	} else if el, err := db.SaveIfAbsent(elem2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(el, elem) {
		t.Error(el)
		t.Fatal(elem)
	}
}
