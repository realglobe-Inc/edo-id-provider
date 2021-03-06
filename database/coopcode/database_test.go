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

package coopcode

import (
	"reflect"
	"testing"
	"time"
)

func testDb(t *testing.T, db Db) {
	if elem, err := db.Get(test_id); err != nil {
		t.Fatal(err)
	} else if elem != nil {
		t.Fatal(elem)
	}

	exp := time.Now().Add(time.Minute)
	tokExp := time.Now().Add(time.Minute)
	elem := New(test_id, exp, test_acnt, test_srcTok, test_scop, tokExp, test_relAcnts, test_frTa, test_toTa)
	saveExp := exp.Add(time.Minute)

	if err := db.Save(elem, saveExp); err != nil {
		t.Fatal(err)
	}

	elem2, err := db.Get(elem.Id())
	if err != nil {
		t.Fatal(err)
	} else if elem2 == nil {
		t.Fatal("no element")
	} else if !reflect.DeepEqual(elem2, elem) {
		t.Error(elem2)
		t.Fatal(elem)
	}

	savedDate := elem2.Date()

	// 確実に時刻を変えるため。
	time.Sleep(time.Millisecond)

	elem2.SetToken(test_tok)
	if ok, err := db.Replace(elem2, savedDate); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("replacement failed")
	} else if ok, err := db.Replace(elem2, savedDate); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("invalid replacement passed")
	}
}
