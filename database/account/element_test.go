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

const (
	test_id   = "EYClXo4mQKwSgPel"
	test_name = "edo-id-provider-tester"
	test_attr = "pds"
)

var (
	test_auth Authenticator
	test_pds  = map[string]interface{}{"type": "single", "uri": "https://pds.example.org"}
)

func init() {
	var err error
	test_auth, err = GenerateStr43Authenticator(test_passwd43, 20)
	if err != nil {
		panic(err)
	}
}

func testElement(t *testing.T, elem Element) {
	if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Error(test_id)
	} else if elem.Name() != test_name {
		t.Error(elem.Name())
		t.Error(test_name)
	} else if !reflect.DeepEqual(elem.Authenticator(), test_auth) {
		t.Error(elem.Authenticator())
		t.Error(test_pds)
	} else if !reflect.DeepEqual(elem.Attribute(test_attr), test_pds) {
		t.Error(elem.Attribute(test_attr))
		t.Error(test_pds)
	}
}
