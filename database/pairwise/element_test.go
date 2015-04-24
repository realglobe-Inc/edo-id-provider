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
	test_acnt    = "EYClXo4mQKwSgPel"
	test_ta      = "https://ta.example.org"
	test_pw_acnt = "i67vMGrkZyyOF1eXNIjekXIG3_iWiFsDsIlKp1-istk"
)

func TestElement(t *testing.T) {
	elem := New(test_acnt, test_ta, test_pw_acnt)
	if elem.Account() != test_acnt {
		t.Error(elem.Account())
		t.Error(test_acnt)
	} else if elem.Ta() != test_ta {
		t.Error(elem.Ta())
		t.Error(test_ta)
	} else if elem.PairwiseAccount() != test_pw_acnt {
		t.Error(elem.PairwiseAccount())
		t.Error(test_pw_acnt)
	} else if elem2 := Generate(test_acnt, test_ta); !reflect.DeepEqual(elem2, elem) {
		t.Error(elem2)
		t.Error(elem)
	}
}
