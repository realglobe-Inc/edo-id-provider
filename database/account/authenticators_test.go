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
	"testing"
)

func TestAuthenticatorFromMapPbkdf2(t *testing.T) {
	m := map[string]interface{}{
		"algorithm": test_alg,
		"salt":      base64url.EncodeToString(test_salt),
		"hash":      base64url.EncodeToString(test_pbkdf2Hash),
	}

	auth, err := authenticatorFromMap(m)
	if err != nil {
		t.Fatal(err)
	} else if err := auth.Verify(test_passwd); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateAuthenticatorPbkdf2(t *testing.T) {
	auth, err := GenerateAuthenticator(test_alg, test_salt, test_passwd)
	if err != nil {
		t.Fatal(err)
	} else if err := auth.Verify(test_passwd); err != nil {
		t.Fatal(err)
	}
}
