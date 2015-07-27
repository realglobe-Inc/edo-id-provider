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

package account

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
)

const (
	test_acntId     = "EYClXo4mQKwSgPel"
	test_acntName   = "edo-id-provider-tester"
	test_acntPasswd = "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
	test_email      = "tester@example.org"

	test_tokId = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
)

var (
	test_acntAuth, _ = account.GenerateAuthenticator("pbkdf2:sha256:1000", []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, test_acntPasswd)
	test_acntAttrs   = map[string]interface{}{
		"email": test_email,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}

	test_ta = tadb.New("https://ta.example.org", nil, nil, nil, false, "")
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
