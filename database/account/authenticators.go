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
)

const (
	AuthTypeStr43 = "STR43"
)

const (
	tagType = "type"
	tagSalt = "salt"
	tagHash = "hash"
)

// json.Unmarshal した結果から復元する。
func authenticatorFromMap(m map[string]interface{}) (Authenticator, error) {
	switch authType, _ := m[tagType].(string); authType {
	case AuthTypeStr43:
		return str43AuthenticatorFromMap(m)
	default:
		return nil, erro.New("unsupported authenticator type " + authType)
	}
}

// {
//     "salt": <ソルト値を base64Url エンコードした文字列>,
//     "hash": <ハッッシュ値を base64Url エンコードした文字列>
// }
func str43AuthenticatorFromMap(m map[string]interface{}) (auth Authenticator, err error) {
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
	return newStr43Authenticator(salt, hVal), nil
}
