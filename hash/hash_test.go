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

package hash

import (
	"crypto/sha256"
	"testing"
)

func TestStringSize(t *testing.T) {
	for _, alg := range []string{"SHA256", "SHA384", "SHA512"} {
		if size := Size(alg); size == 0 {
			t.Fatal(alg)
		}
	}
	if size := Size("unknown"); size != 0 {
		t.Fatal("no error")
	}
}

func TestGenerator(t *testing.T) {
	for _, alg := range []string{"SHA256", "SHA384", "SHA512"} {
		if hGen := Generator(alg); hGen == 0 {
			t.Fatal(alg)
		}
	}
	if hGen := Generator("unknown"); hGen != 0 {
		t.Fatal("no error")
	}
}

func TestHashing(t *testing.T) {
	h := "sRIadA8LhMSb7MUizbTGXA"
	h2 := Hashing(sha256.New(), []byte("eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOlsiaHR0cHM6Ly9pZHAyLmV4YW1wbGUub3JnIl0sImV4cCI6MTQyNTQ1MjgzNSwiaGFzaF9hbGciOiJTSEEyNTYiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsImp0aSI6InlHLTh4Zm1Pb1Q3RDRERE0iLCJyZWxhdGVkX3VzZXJzIjp7Im9ic2VydmVyIjoiR2JyOTNrdnRIUGt1VVg0WWxSRDRRQSJ9LCJzdWIiOiJodHRwczovL2Zyb20uZXhhbXBsZS5vcmciLCJ0b19jbGllbnQiOiJodHRwczovL3RvLmV4YW1wbGUub3JnIn0.YjH8_n8D00qjIDtbBQW1JpkxBoDFs78Eepo0Jn1WeI8PjTdDCiwZy8ZcvOfJTisoEFPunjWYplVE7wqUUV9mxw"))
	if h2 != h {
		t.Error(h)
		t.Fatal(h2)
	}
}
