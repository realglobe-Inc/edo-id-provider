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

package key

import (
	"reflect"
	"testing"

	"github.com/realglobe-Inc/edo-lib/jwk"
)

var (
	test_key, _ = jwk.FromMap(map[string]interface{}{
		"kty": "RSA",
		"n":   "5cadP6Vvv6ABglXpSeXYxPB321gtSwmjccsHr2-YKmBm22KWF2A1b68LJ3mA8eG5NPSRL6macCMttxsoAKwaCxOxn-6dNOKXNLQ1S0WsE4yY2QLoi9Cj_sY8yfdk_wb0ZM5kyE99GjFFLDvnh-RjHIf2cbXPyPfbeLigeeon7jsxOw",
		"e":   "AQAB",
		"d":   "gOV1-Oo5UenUbuT6xXWmsHOlCOriHaH-iis22HdliQAjMxaO0_Yog8pSG4bRit7xIn-_olkmRZm2X21gd2AUC_mkE7Nytw5t_pioMzupEVVGApIFuc2_ryf5VPSznx3zk5FY6XCgUf6BnJ188WRUv3CnnNuAmEJtP6MhWmoKlPMpgQ",
		"p":   "9556qFgzilKEEhQ41fVzvLm5vKpiCc0IABG1CDQ_VTr4KGoOcqSHx6__yqYFQlzgizkG-zVxBQSs-6GZ3eA-t4s",
		"q":   "7Y2H2tRgIm9UjN0OlszOBcOXqPicE5KlseuNCIJZo1SyW30h-N2ssjCeiSDPrqm5QGZ637EAmhvNsPNOxzwLIxE",
		"dp":  "sa3EMdvoT87Z-ecMyWpw-_EA-AICiynWHcaW8iYbc9r2inlfmJ-61mzRzOXITFA8x2nKOqOkT4eFYKIauHzaQ_U",
		"dq":  "EhoQ2ioI0VbueHV34SHmKSZIbkXTjuJD4hTzAEz-i6Wuma4lYpNxz3pI-mYXrVWdmjy07ErOou-vcuZ3gFMg_iE",
		"qi":  "DbisQAteFbdCaNy6TyNy5UgZjdPba1bhKI3iIXalno_5HRrK4tUzu9VHdYVj5-iscIw5za9cPMLFr3zQvWa-gzA",
	})
	test_sigKey, _ = jwk.FromMap(map[string]interface{}{
		"kid": "1",
		"use": "sig",
		"kty": "EC",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"crv": "P-256",
	})
	test_encKey, _ = jwk.FromMap(map[string]interface{}{
		"kid": "2",
		"use": "enc",
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_veriKey, _ = jwk.FromMap(map[string]interface{}{
		"kid":     "3",
		"key_ops": []interface{}{"verify"},
		"kty":     "EC",
		"crv":     "P-521",
		"x":       "AXA9Y2pY1g_Cs4W6Nto7ebjDKOsaRHxo6EYRWjk1XHZaA7HnkeHg13x24OWHelqdZiuo7J1VbRKJ4ohZPjKX-AL7",
		"y":       "AeC5zdcUHLDdhQAvdsnnwD8rgNWjMdlWsqXZxv7Ar7ly5xmZGxEDtcJuhfhn8R9PXeScPH2soF3dFYPCuDkF4Gns",
		"d":       "AK9ejP0HuUt7ojjI9p20986DGqG-5jc9UWMtnMqxNvIvBScTJflS2lE-6CRsJZKk6ChWI6U4ahXDH0cCCFWTAvI_",
	})
)

// test_key, test_sigKey, test_encKey, test_veriKey, が保存されていることが前提。
func testDb(t *testing.T, db Db) {
	keys, err := db.Get()
	if err != nil {
		t.Fatal(err)
	} else if len(keys) != 4 {
		t.Fatal(keys)
	}
	for _, key := range keys {
		if !deepEqualToOne(key, test_key, test_sigKey, test_encKey, test_veriKey) {
			t.Error(key)
			t.Error(test_key, test_sigKey, test_encKey, test_veriKey)
		}
	}
}

func deepEqualToOne(a interface{}, targets ...interface{}) bool {
	for _, target := range targets {
		if reflect.DeepEqual(a, target) {
			return true
		}
	}
	return false
}
