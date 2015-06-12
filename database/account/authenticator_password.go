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
	"github.com/realglobe-Inc/edo-lib/password"
	"github.com/realglobe-Inc/go-lib/erro"
)

// ソルトでハッシュ値化したパスワード認証。
type passwordAuthenticator struct {
	alg  string
	salt []byte

	hVal []byte
}

func newPasswordAuthenticator(alg string, salt, hVal []byte) *passwordAuthenticator {
	return &passwordAuthenticator{alg, salt, hVal}
}

func (this *passwordAuthenticator) Verify(params ...interface{}) error {
	if len(params) < 1 {
		return erro.New("password required")
	}
	passwd, ok := params[0].(string)
	if !ok {
		return erro.New("invalid password")
	}
	hVal, err := password.Calculate(this.alg, this.salt, passwd)
	if err != nil {
		return erro.Wrap(err)
	} else if !bytes.Equal(hVal, this.hVal) {
		return erro.New("verification failed")
	}
	return nil
}

// パスワードからつくる。
func generatePasswordAuthenticator(alg string, salt []byte, passwd string) (Authenticator, error) {
	hVal, err := password.Calculate(alg, salt, passwd)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return newPasswordAuthenticator(alg, salt, hVal), nil
}
