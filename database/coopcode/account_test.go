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
	"testing"
)

const (
	test_acntId  = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acntTag = "self"
)

func TestAccount(t *testing.T) {
	acnt := NewAccount(test_acntId, test_acntTag)

	if acnt.Id() != test_acntId {
		t.Error(acnt.Id())
		t.Fatal(test_acntId)
	} else if acnt.Tag() != test_acntTag {
		t.Error(acnt.Tag())
		t.Fatal(test_acntTag)
	}
}
