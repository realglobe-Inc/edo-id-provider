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

func testDb(t *testing.T, db Db) {
	if elem, err := db.GetByPairwise(test_ta, test_pw_acnt); err != nil {
		t.Error(err)
		return
	} else if elem != nil {
		t.Error(elem)
		return
	}

	elem := New(test_acnt, test_ta, test_pw_acnt)
	if err := db.Save(elem); err != nil {
		t.Error(err)
		return
	} else if elem2, err := db.GetByPairwise(elem.Ta(), elem.PairwiseAccount()); err != nil {
		t.Error(err)
		return
	} else if elem2 == nil {
		t.Error("no element")
		return
	} else if !reflect.DeepEqual(elem2, elem) {
		t.Error(elem2)
		t.Error(elem)
		return
	}
}
