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

func testElement(t *testing.T, elem Element) {
	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if elem.Name() != test_name {
		t.Error(elem.Name())
		t.Fatal(test_name)
	} else if !reflect.DeepEqual(elem.Authenticator(), test_auth) {
		t.Error(elem.Authenticator())
		t.Fatal(test_pds)
	} else if !reflect.DeepEqual(elem.Attribute(test_attr), test_pds) {
		t.Error(elem.Attribute(test_attr))
		t.Fatal(test_pds)
	}

	elem.SetAttribute(test_attr+"a", "abcde")
	if !reflect.DeepEqual(elem.Attribute(test_attr+"a"), "abcde") {
		t.Error(elem.Attribute(test_attr + "a"))
		t.Fatal("abcde")
	}
}
