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

// OpenID Connect 1.0 や EDO に関わるハッシュ関数周り。
package hash

import (
	"crypto"
	"encoding/base64"
	"hash"

	hashutil "github.com/realglobe-Inc/edo-lib/hash"
)

// related_users に入れるハッシュ値の文字列としての長さを返す。
// 該当するものが無い場合は 0。
func Size(alg string) int {
	switch alg {
	case tagSha256:
		return (256/2 + 5) / 6
	case tagSha384:
		return (384/2 + 5) / 6
	case tagSha512:
		return (512/2 + 5) / 6
	default:
		return 0
	}
}

// 該当するものが無い場合は 0。
func Generator(alg string) crypto.Hash {
	switch alg {
	case tagSha256:
		return crypto.SHA256
	case tagSha384:
		return crypto.SHA384
	case tagSha512:
		return crypto.SHA512
	default:
		return 0
	}
}

// ハッシュ値を計算して前半分を Base64URL エンコードして返す。
func Hashing(hFun hash.Hash, data ...[]byte) string {
	sum := hashutil.Hashing(hFun, data...)
	return base64.RawURLEncoding.EncodeToString(sum[:len(sum)/2])
}
