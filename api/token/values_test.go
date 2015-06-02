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

package token

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	test_codId = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"

	test_acntId     = "EYClXo4mQKwSgPel"
	test_acntName   = "edo-id-provider-tester"
	test_acntPasswd = "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
	test_email      = "tester@example.org"

	test_taSigAlg = "ES384"
	test_rediUri  = "https://ta.example.org/callback"
	test_nonc     = "Wjj1_YUOlR"
	test_jti      = "R-seIeMPBly4xPAh"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})

	test_acntAuth, _ = account.GenerateStr43Authenticator(test_acntPasswd, 20)
	test_acntAttrs   = map[string]interface{}{
		"email": test_email,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}
	test_taKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_ta = tadb.New("https://ta.example.org", nil, strsetutil.New(test_rediUri), []jwk.Key{test_taKey}, false, "")
)

func newTestAccount() account.Element {
	return account.New(test_acntId, test_acntName, test_acntAuth, clone(test_acntAttrs))
}

// 1 段目だけのコピー。
func clone(m map[string]interface{}) map[string]interface{} {
	m2 := map[string]interface{}{}
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func newCode() *authcode.Element {
	now := time.Now()
	return authcode.New(test_codId, now.Add(time.Minute), test_acntId, now, strsetutil.New("openid"),
		nil, strsetutil.New("email"), test_ta.Id(), test_rediUri, test_nonc)
}

func newRequest(codId, aud string) (*http.Request, error) {
	q := url.Values{}
	q.Set("grant_type", "authorization_code")
	q.Set("code", codId)
	q.Set("redirect_uri", test_rediUri)
	q.Set("client_id", test_ta.Id())
	q.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	{
		jt := jwt.New()
		jt.SetHeader("alg", test_taSigAlg)
		jt.SetClaim("iss", test_ta.Id())
		jt.SetClaim("sub", test_ta.Id())
		jt.SetClaim("aud", audience.New(aud))
		jt.SetClaim("jti", test_jti)
		now := time.Now()
		jt.SetClaim("exp", now.Add(time.Minute).Unix())
		jt.SetClaim("iat", now.Unix())
		if err := jt.Sign(test_ta.Keys()); err != nil {
			return nil, erro.Wrap(err)
		}
		buff, err := jt.Encode()
		if err != nil {
			return nil, erro.Wrap(err)
		}
		q.Set("client_assertion", string(buff))
	}
	r, err := http.NewRequest("POST", aud, strings.NewReader(q.Encode()))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r, nil
}
