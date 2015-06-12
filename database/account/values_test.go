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
	_ "crypto/sha256"
)

const (
	test_alg    = "pbkdf2:sha256:1000"
	test_attr   = "pds"
	test_id     = "EYClXo4mQKwSgPel"
	test_name   = "edo-id-provider-tester"
	test_passwd = "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
)

var (
	test_pds  = map[string]interface{}{"type": "single", "uri": "https://pds.example.org"}
	test_salt = []byte{
		0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19,
	}
	test_pbkdf2Hash = []byte{
		49, 6, 104, 72, 138, 241, 200, 28,
		29, 72, 81, 87, 91, 120, 30, 102,
		152, 193, 0, 152, 180, 168, 250, 197,
		71, 88, 124, 224, 48, 207, 15, 143,
	}
	test_auth = newPasswordAuthenticator(test_alg, test_salt, test_pbkdf2Hash)
	test_elem = newElement(test_id, test_name, test_auth, map[string]interface{}{test_attr: test_pds})
)
