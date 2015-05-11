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

const (
	test_iss = "https://ta.example.org"
	test_id  = "R-seIeMPBly4xPAh"
)

func TestElement(t *testing.T) {
	exp := time.Now().Add(time.Second)

	if elem := New(test_iss, test_id, exp); elem.Issuer() != test_iss {
		t.Error(elem.Issuer())
		t.Fatal(test_iss)
	} else if elem.Id() != test_id {
		t.Error(elem.Id())
		t.Fatal(test_id)
	} else if !elem.Expires().Equal(exp) {
		t.Error(elem.Expires())
		t.Fatal(exp)
	}
}
