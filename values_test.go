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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-lib/jwk"
)

const (
	test_label = "edo-test"
)

const (
	test_pathOk     = "/ok"
	test_pathAuth   = "/auth"
	test_pathSel    = "/auth/select"
	test_pathLgin   = "/auth/login"
	test_pathCons   = "/auth/consent"
	test_pathTa     = "/api/info/ta"
	test_pathTok    = "/api/token"
	test_pathAcnt   = "/api/info/account"
	test_pathCoopFr = "/api/coop/from"
	test_pathCoopTo = "/api/coop/to"

	test_pathUi     = "/ui"
	test_pathSelUi  = "/ui/select.html"
	test_pathLginUi = "/ui/login.html"
	test_pathConsUi = "/ui/consent.html"

	test_taPathCb = "/callback"
)

var (
	test_idpPriKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})
	test_idpPubKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
	})
	test_taPriKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_taPubKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
	})
)

const (
	test_taName   = "The TA"
	test_taNameJa = "かの TA"
)

const (
	test_acntId     = "EYClXo4mQKwSgPel"
	test_acntName   = "edo-id-provider-tester"
	test_acntPasswd = "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
)

var (
	test_acntAuth, _ = account.GenerateStr43Authenticator(test_acntPasswd, 20)
	test_acntAttrs   = map[string]interface{}{
		"email": "tester@example.org",
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}
)
