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
	"bytes"
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

const (
	test_tokId = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"

	test_acntTag   = "main-user"
	test_acntId    = "EYClXo4mQKwSgPel"
	test_acntEmail = "tester@example.org"

	test_subAcnt1Tag   = "sub-user1"
	test_subAcnt1Id    = "U7pdvT8dYbBFWXdc"
	test_subAcnt1Email = "subtester1@example.org"

	test_frTaSigAlg = "ES384"
	test_jti        = "R-seIeMPBly4xPAh"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})

	test_acntAttrs = map[string]interface{}{
		"email": test_acntEmail,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}
	test_subAcnt1Attrs = map[string]interface{}{
		"email": test_subAcnt1Email,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}

	test_frTaKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_frTa = tadb.New("https://from.example.org", nil, nil, []jwk.Key{test_frTaKey}, false, "")
	test_toTa = tadb.New("https://to.example.org", nil, nil, nil, false, "")
)

func newTestMainAccount() account.Element {
	return account.New(test_acntId, "", nil, clone(test_acntAttrs))
}

func newTestSubAccount1() account.Element {
	return account.New(test_subAcnt1Id, "", nil, clone(test_subAcnt1Attrs))
}

// 1 段目だけのコピー。
func clone(m map[string]interface{}) map[string]interface{} {
	m2 := map[string]interface{}{}
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func newTestSingleRequest(aud string) (*http.Request, error) {
	m := map[string]interface{}{
		"response_type":         "code_token",
		"from_client":           test_frTa.Id(),
		"to_client":             test_toTa.Id(),
		"grant_type":            "access_token",
		"access_token":          test_tokId,
		"user_tag":              test_acntTag,
		"users":                 map[string]string{test_subAcnt1Tag: test_subAcnt1Id},
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}
	{
		jt := jwt.New()
		jt.SetHeader("alg", test_frTaSigAlg)
		jt.SetClaim("iss", test_frTa.Id())
		jt.SetClaim("sub", test_frTa.Id())
		jt.SetClaim("aud", audience.New(aud))
		jt.SetClaim("jti", test_jti)
		now := time.Now()
		jt.SetClaim("exp", now.Add(time.Minute).Unix())
		jt.SetClaim("iat", now.Unix())
		if err := jt.Sign(test_frTa.Keys()); err != nil {
			return nil, erro.Wrap(err)
		}
		buff, err := jt.Encode()
		if err != nil {
			return nil, erro.Wrap(err)
		}
		m["client_assertion"] = string(buff)
	}
	body, err := json.Marshal(m)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r, err := http.NewRequest("POST", aud, bytes.NewReader(body))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("Content-Type", "application/json")
	return r, nil
}
