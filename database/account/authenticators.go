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
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/go-lib/erro"
	"strings"
)

// json.Unmarshal した結果から復元する。
func authenticatorFromMap(m map[string]interface{}) (auth Authenticator, err error) {
	alg, _ := m[tagAlgorithm].(string)
	parts := strings.SplitN(alg, algSep, 2)
	if len(parts) == 0 {
		return nil, erro.New("no algorithm")
	}

	switch parts[0] {
	case tagPbkdf2:
		// パスワード認証。
		var salt []byte
		if s, _ := m[tagSalt].(string); s == "" {
			return nil, erro.New("no salt")
		} else if salt, err = base64url.DecodeString(s); err != nil {
			return nil, erro.Wrap(err)
		}
		var hVal []byte
		if s, _ := m[tagHash].(string); s == "" {
			return nil, erro.New("no hash value")
		} else if hVal, err = base64url.DecodeString(s); err != nil {
			return nil, erro.Wrap(err)
		}
		return newPasswordAuthenticator(alg, salt, hVal), nil
	default:
		return nil, erro.New("unsupported authenticator algorithm " + parts[0])
	}
}

func GenerateAuthenticator(alg string, params ...interface{}) (Authenticator, error) {
	parts := strings.SplitN(alg, algSep, 2)
	if len(parts) == 0 {
		return nil, erro.New("no authenticator algorithm")
	}

	switch parts[0] {
	case tagPbkdf2:
		if len(params) < 1 {
			return nil, erro.New("no " + parts[0] + " salt")
		} else if len(params) < 2 {
			return nil, erro.New("no " + parts[0] + " password")
		}

		salt, _ := params[0].([]byte)
		if salt == nil {
			return nil, erro.New("invalid salt")
		}
		passwd, _ := params[1].(string)
		if passwd == "" {
			return nil, erro.New("invlid password")
		}
		return generatePasswordAuthenticator(alg, salt, passwd)
	default:
		return nil, erro.New("unsupported algorithm " + parts[0])
	}
}
