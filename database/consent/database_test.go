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

package consent

import (
	"reflect"
	"testing"
)

func testDb(t *testing.T, db Db) {
	if el, err := db.Get(test_acnt, test_ta); err != nil {
		t.Fatal(err)
	} else if el != nil {
		t.Fatal(el)
	}

	elem := New(test_acnt, test_ta)
	elem.Scope().SetAllow(test_scop)

	if err := db.Save(elem); err != nil {
		t.Fatal(err)
	}

	elem2, err := db.Get(elem.Account(), elem.Ta())
	if err != nil {
		t.Fatal(err)
	} else if elem2 == nil {
		t.Fatal("no element")
	} else if !reflect.DeepEqual(elem2, elem) {
		t.Error(elem2)
		t.Fatal(elem)
	}
}
