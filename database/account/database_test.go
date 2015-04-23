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

var test_elem = newElement(test_id, test_name, test_auth, map[string]interface{}{test_attr: test_pds})

// test_elem が保存されていることが前提。
func testDb(t *testing.T, db Db) {
	if elem, err := db.Get(test_id); err != nil {
		t.Error(err)
	} else if elem == nil {
		t.Error("no element")
	} else if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Error(test_id)
	} else if elem.Name() != test_name {
		t.Error(elem.Name())
		t.Error(test_name)
	} else if !reflect.DeepEqual(elem.Attribute(test_attr), test_pds) {
		t.Error(elem.Attribute(test_attr))
		t.Error(test_pds)
	}
}
