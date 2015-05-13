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

package request

import (
	"net/http"
	"testing"
)

func TestRequest(t *testing.T) {
	id := "XAOiyqgngWGzZbgl6j1w6Zm3ytHeI-"

	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:   "SID",
		Value:  id,
		Path:   "/",
		MaxAge: 7 * 24 * 3600,
	})

	r.RemoteAddr = "192.168.0.18:55555"
	if req := Parse(r, "SID"); req.Session() != id {
		t.Error(req.Session())
		t.Fatal(id)
	} else if req.Source() != "192.168.0.18:55555" {
		t.Error(req.Source())
		t.Fatal("192.168.0.18:55555")
	} else if req.String() != "XAOiyqgn@192.168.0.18:55555" {
		t.Error(req.String())
		t.Fatal("XAOiyqgn@192.168.0.18:55555")
	}

	r.Header.Set("X-Forwarded-For", "203.0.113.34, 192.168.0.12")
	if req := Parse(r, "SID"); req.Session() != id {
		t.Error(req.Session())
		t.Fatal(id)
	} else if req.Source() != "203.0.113.34" {
		t.Error(req.Source())
		t.Fatal("203.0.113.34")
	} else if req.String() != "XAOiyqgn@203.0.113.34" {
		t.Error(req.String())
		t.Fatal("XAOiyqgn@203.0.113.34")
	}
}
