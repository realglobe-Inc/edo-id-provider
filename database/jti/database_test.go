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

package jti

import (
	"testing"
	"time"
)

func testDb(t *testing.T, db Db) {
	elem := New(test_iss, test_id, time.Now().Add(time.Minute))
	if ok, err := db.SaveIfAbsent(elem); err != nil {
		t.Error(err)
		return
	} else if !ok {
		t.Error("saving failed")
		return
	}

	elem2 := New(test_iss, test_id, time.Now().Add(time.Minute))
	if ok, err := db.SaveIfAbsent(elem2); err != nil {
		t.Error(err)
		return
	} else if ok {
		t.Error("saving passed")
		return
	}
}
