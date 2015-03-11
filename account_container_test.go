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
	"testing"
	"time"
)

var testAcc = newAccount(map[string]interface{}{
	"id":        "testaccount",
	"username":  "testaccountname",
	"password":  "testaccountpassword",
	"update_at": time.Now(),

	"email": "testaccount@example.org",
})

func testAccountContainer(t *testing.T, accCont accountContainer) {
	defer accCont.close()

	if acc, err := accCont.get(testAcc.id()); err != nil {
		t.Fatal(err)
	} else if acc.id() != testAcc.id() {
		t.Error(acc)
	} else if acc.name() != testAcc.name() {
		t.Error(acc)
	} else if acc.password() != testAcc.password() {
		t.Error(acc)
	}

	if acc, err := accCont.get(testAcc.id() + "a"); err != nil {
		t.Fatal(err)
	} else if acc != nil {
		t.Error(acc)
	}

	if acc, err := accCont.getByName(testAcc.name()); err != nil {
		t.Fatal(err)
	} else if acc.id() != testAcc.id() {
		t.Error(acc)
	} else if acc.name() != testAcc.name() {
		t.Error(acc)
	} else if acc.password() != testAcc.password() {
		t.Error(acc)
	}

	if acc, err := accCont.getByName(testAcc.name() + "a"); err != nil {
		t.Fatal(err)
	} else if acc != nil {
		t.Error(acc)
	}
}
