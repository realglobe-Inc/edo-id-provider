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
	"bytes"
	"testing"
)

const (
	test_id = "https://ta.example.org"
)

var (
	test_salt = []byte{0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19,
	}
)

func TestElement(t *testing.T) {
	if elem := New(test_id, test_salt); elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if !bytes.Equal(elem.Salt(), test_salt) {
		t.Error(elem.Salt())
		t.Fatal(test_salt)
	}
}
