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
	test_acnt_id  = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acnt_tag = "self"
)

func TestAccount(t *testing.T) {
	acnt := NewAccount(test_acnt_id, test_acnt_tag)

	if acnt.Id() != test_acnt_id {
		t.Error(acnt.Id())
		t.Error(test_acnt_id)
	} else if acnt.Tag() != test_acnt_tag {
		t.Error(acnt.Tag())
		t.Error(test_acnt_tag)
	}
}
