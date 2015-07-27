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

package account

import (
	"reflect"
	"testing"
)

// test_elem が保存されていることが前提。
func testDb(t *testing.T, db Db) {
	if elem, err := db.Get(test_elem.Id() + "a"); err != nil {
		t.Fatal(err)
	} else if elem != nil {
		t.Fatal(elem)
	} else if elem, err := db.Get(test_elem.Id()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(elem, test_elem) {
		t.Error(elem)
		t.Fatal(test_elem)
	} else if elem, err := db.GetByName(test_elem.Name() + "a"); err != nil {
		t.Fatal(err)
	} else if elem != nil {
		t.Fatal(elem)
	} else if elem, err := db.GetByName(test_elem.Name()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(elem, test_elem) {
		t.Error(elem)
		t.Fatal(test_elem)
	}
}
