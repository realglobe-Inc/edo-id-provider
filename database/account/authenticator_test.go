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
	"testing"
)

func testAuthenticatorType(t *testing.T, auth Authenticator, typ string) {
	if auth.Type() != typ {
		t.Error(auth.Type())
		t.Error(typ)
	}
}

func testAuthenticatorVerify(t *testing.T, auth Authenticator, passwd string, params ...interface{}) {
	if !auth.Verify(passwd, params...) {
		t.Error("verification error")
	} else if auth.Verify(passwd[1:]+passwd[:1], params...) {
		t.Error("security error")
	}
}
