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

package assertion

import (
	"github.com/realglobe-Inc/edo-lib/jwk"
	"testing"
	"time"
)

func TestAssertion(t *testing.T) {
	raw := []byte("eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZy9hcGkvdG9rZW4iLCJleHAiOjEwMDAwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly90YS5leGFtcGxlLm9yZyIsImp0aSI6IjVDWXAxUFFvMm1UV3Ztak8wcUFQIiwic3ViIjoiaHR0cHM6Ly90YS5leGFtcGxlLm9yZyJ9.EqiV-a-hrMDZGkWwJdCYSoOeQUrsJVW4hxMic2W5YF4MOT0_4VJQFkamgpZJHgK82RcThuWiJ4iAh8dz8tsNJA")
	taId := "https://ta.example.org"
	key, err := jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "BuFKZIt6O4s3zNwBvoEOQ6yHqiD1ovhpw-W7Kdtqu9U",
		"y":   "HuyOY5osQFSBYj8TN-ctJF8v5IP8NYoeLdkDb-lSjDw",
	})
	if err != nil {
		t.Fatal(err)
	}
	iss := "https://ta.example.org"
	aud := "https://idp.example.org/api/token"
	id := "5CYp1PQo2mTWvmjO0qAP"
	exp := time.Unix(100000000000, 0)

	ass, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	} else if err := ass.Verify(taId, []jwk.Key{key}, aud); err != nil {
		t.Fatal(err)
	} else if ass.Issuer() != iss {
		t.Error(ass.Issuer())
		t.Fatal(iss)
	} else if ass.Id() != id {
		t.Error(ass.Id())
		t.Fatal(id)
	} else if !ass.Expires().Equal(exp) {
		t.Error(ass.Expires())
		t.Fatal(exp)
	}
}
