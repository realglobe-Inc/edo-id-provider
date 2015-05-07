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

func TestBaseRequest(t *testing.T) {
	id := "XAOiyqgngWGzZbgl6j1w6Zm3ytHeI-"

	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:   "Id-Provider",
		Value:  id,
		Path:   "/",
		MaxAge: 7 * 24 * 3600,
	})

	r.RemoteAddr = "192.168.0.18:55555"
	if req := newBaseRequest(r); req.session() != id {
		t.Error(req.session())
		t.Fatal(id)
	} else if req.source() != "192.168.0.18:55555" {
		t.Error(req.source())
		t.Fatal("192.168.0.18:55555")
	}

	r.Header.Set("X-Forwarded-For", "203.0.113.34, 192.168.0.12")
	if req := newBaseRequest(r); req.session() != id {
		t.Error(req.session())
		t.Fatal(id)
	} else if req.source() != "203.0.113.34" {
		t.Error(r)
		t.Error(req.source())
		t.Fatal("203.0.113.34")
	}
}
