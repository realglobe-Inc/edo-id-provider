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
	"net/http"
	"testing"
)

func TestAccountRequest(t *testing.T) {
	token := "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"

	r, err := http.NewRequest("GET", "https://idp.example.org/api/info/account", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer "+token)

	req := newAccountRequest(r)
	if req.scheme() != "Bearer" {
		t.Error(req.scheme())
		t.Fatal("Bearer")
	} else if req.token() != token {
		t.Error(req.token())
		t.Fatal(token)
	}
}
