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

package coopfrom

import (
	"reflect"
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
)

func TestReferral(t *testing.T) {
	raw := []byte("eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOlsiaHR0cHM6Ly9pZHAyLmV4YW1wbGUub3JnIl0sImV4cCI6MTQyNTQ1MjgzNSwiaGFzaF9hbGciOiJTSEEyNTYiLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsImp0aSI6InlHLTh4Zm1Pb1Q3RDRERE0iLCJyZWxhdGVkX3VzZXJzIjp7Im9ic2VydmVyIjoiR2JyOTNrdnRIUGt1VVg0WWxSRDRRQSJ9LCJzdWIiOiJodHRwczovL2Zyb20uZXhhbXBsZS5vcmciLCJ0b19jbGllbnQiOiJodHRwczovL3RvLmV4YW1wbGUub3JnIn0.YjH8_n8D00qjIDtbBQW1JpkxBoDFs78Eepo0Jn1WeI8PjTdDCiwZy8ZcvOfJTisoEFPunjWYplVE7wqUUV9mxw")

	ref, err := parseReferral(raw)
	if err != nil {
		t.Fatal(err)
	} else if id := "yG-8xfmOoT7D4DDM"; ref.id() != id {
		t.Error(ref.id())
		t.Fatal(id)
	} else if idp := "https://idp.example.org"; ref.idProvider() != idp {
		t.Error(ref.idProvider())
		t.Fatal(idp)
	} else if frTa := "https://from.example.org"; ref.fromTa() != frTa {
		t.Error(ref.fromTa())
		t.Fatal(frTa)
	} else if toTa := "https://to.example.org"; ref.toTa() != toTa {
		t.Error(ref.toTa())
		t.Fatal(toTa)
	} else if aud := strsetutil.New("https://idp2.example.org"); !reflect.DeepEqual(ref.audience(), aud) {
		t.Error(ref.audience())
		t.Fatal(aud)
	} else if relAcnts := map[string]string{"observer": "Gbr93kvtHPkuUX4YlRD4QA"}; !reflect.DeepEqual(ref.relatedAccounts(), relAcnts) {
		t.Error(ref.relatedAccounts)
		t.Fatal(relAcnts)
	} else if hAlg := "SHA256"; ref.hashAlgorithm() != hAlg {
		t.Error(ref.hashAlgorithm())
		t.Fatal(hAlg)
	} else if exp := time.Unix(1425452835, 0); !ref.expires().Equal(exp) {
		t.Error(ref.expires())
		t.Fatal(exp)
	}
}
