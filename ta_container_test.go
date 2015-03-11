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

package main

import (
	"reflect"
	"testing"
)

func testTaContainer(t *testing.T, taCont taContainer) {
	defer taCont.close()

	if ta_, err := taCont.get(testTa.id()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ta_, testTa) {
		t.Error(ta_, testTa)
	}

	if ta_, err := taCont.get(testTa.id() + "a"); err != nil {
		t.Fatal(err)
	} else if ta_ != nil {
		t.Error(ta_)
	}
}
