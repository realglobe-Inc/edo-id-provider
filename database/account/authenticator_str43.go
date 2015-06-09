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
	"bytes"
	"crypto/sha256"
	"github.com/realglobe-Inc/edo-lib/hash"
	"github.com/realglobe-Inc/edo-lib/secrand"
	"github.com/realglobe-Inc/go-lib/erro"
)

// 43 文字のパスワードを受け入れる認証機構。
type str43Authenticator struct {
	salt []byte
	hVal []byte
}

func newStr43Authenticator(salt, hVal []byte) *str43Authenticator {
	return &str43Authenticator{salt, hVal}
}

func (this *str43Authenticator) Type() string {
	return AuthTypeStr43
}

func (this *str43Authenticator) Verify(passwd string, params ...interface{}) bool {
	if len(passwd) != 43 {
		return false
	}
	return bytes.Equal(hash.Hashing(sha256.New(), this.salt, []byte{0}, []byte(passwd)), this.hVal)
}

// パスワードからつくる。
func GenerateStr43Authenticator(passwd string, sLen int) (Authenticator, error) {
	if len(passwd) != 43 {
		return nil, erro.New("invalid password length")
	}

	salt, err := secrand.Bytes(sLen)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	return newStr43Authenticator(salt, hash.Hashing(sha256.New(), salt, []byte{0}, []byte(passwd))), nil
}
