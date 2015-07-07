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

package auth

import (
	"net/http"
	"net/url"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
)

const (
	test_rediUri = "https://ta.example.org/callback"
	test_passwd  = "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
	test_stat    = "YJgUit_Wx5"
	test_nonc    = "Wjj1_YUOlR"
	test_sessId  = "XAOiyqgngWGzZbgl6j1w6Zm3ytHeI-"
	test_idpId   = "https://idp.example.org"
	test_ticId   = "-TRO_YRa1B"

	test_acntName = "edo-id-provider-tester"
	test_lang     = "ja-JP"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})
	test_acnt = account.New(
		"EYClXo4mQKwSgPel",
		"edo-id-provider-tester",
		func() account.Authenticator {
			auth, err := account.GenerateAuthenticator("pbkdf2:sha256:1000", []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, test_passwd)
			if err != nil {
				panic(err)
			}
			return auth
		}(),
		map[string]interface{}{
			"pds": map[string]interface{}{
				"type": "single",
				"uri":  "https://pds.example.org",
			},
		},
	)
	test_taKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_ta = tadb.New(
		"https://ta.example.org",
		nil,
		strsetutil.New(test_rediUri),
		[]jwk.Key{test_taKey},
		false,
		"",
	)
	test_authQuery = "response_type=code&scope=openid" +
		"&client_id=" + url.QueryEscape(test_ta.Id()) +
		"&redirect_uri=" + url.QueryEscape(test_rediUri) +
		"&state=" + url.QueryEscape(test_stat) +
		"&nonce=" + url.QueryEscape(test_nonc)
	test_req = func() *session.Request {
		r, err := http.NewRequest("GET", test_idpId+"/auth?"+test_authQuery, nil)
		if err != nil {
			panic(err)
		}
		req, err := session.ParseRequest(r)
		if err != nil {
			panic(err)
		}
		return req
	}()
	test_selQuery = "ticket=" + url.QueryEscape(test_ticId) +
		"&username=" + url.QueryEscape(test_acnt.Name())
	test_lginQuery = "ticket=" + url.QueryEscape(test_ticId) +
		"&username=" + url.QueryEscape(test_acnt.Name()) +
		"&pass_type=password" +
		"&password=" + url.QueryEscape(test_passwd)
	test_consQuery = "ticket=" + url.QueryEscape(test_ticId) +
		"&allowed_scope=openid"
)
