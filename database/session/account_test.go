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

package session

import (
	"testing"
)

func TestAccountNew(t *testing.T) {
	a := NewAccount("test-account-id", "test-account-name")
	a.Login()
	if !a.LoggedIn() {
		t.Error("account not logged in")
	}

	b := a.New()

	if b.Id() != a.Id() {
		t.Error(b.Id())
	} else if b.Name() != a.Name() {
		t.Error(b.Name())
	} else if b.LoggedIn() {
		t.Error("new account logged in")
	}
}
