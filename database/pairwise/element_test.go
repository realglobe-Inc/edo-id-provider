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

package pairwise

import (
	"reflect"
	"testing"
)

const (
	test_acnt   = "EYClXo4mQKwSgPel"
	test_sect   = "https://ta.example.org"
	test_pwAcnt = "X4mJU00-YAQUNpuFC1sf6YejJ0ZqdagM3SkNSfOCAq8"
)

var (
	test_salt = []byte{
		0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19,
	}
)

func TestElement(t *testing.T) {
	elem := New(test_acnt, test_sect, test_pwAcnt)
	if elem.Account() != test_acnt {
		t.Error(elem.Account())
		t.Fatal(test_acnt)
	} else if elem.Sector() != test_sect {
		t.Error(elem.Sector())
		t.Fatal(test_sect)
	} else if elem.Pairwise() != test_pwAcnt {
		t.Error(elem.Pairwise())
		t.Fatal(test_pwAcnt)
	} else if elem2 := Generate(test_acnt, test_sect, test_salt); !reflect.DeepEqual(elem2, elem) {
		t.Error(elem2)
		t.Fatal(elem)
	}
}
