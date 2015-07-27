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

package idputil

const (
	// アンダースコア。
	tagAlg  = "alg"
	tagNone = "none"
	tagSub  = "sub"
	tagKid  = "kid"
	tagIss  = "iss"
	tagAud  = "aud"
	tagExp  = "exp"
	tagIat  = "iat"

	// ハイフン。
	tagNo_cache = "no-cache"
	tagNo_store = "no-store"

	// 頭大文字、ハイフン。
	tagPragma        = "Pragma"
	tagCache_control = "Cache-Control"
	tagContent_type  = "Content-Type"
)

const (
	// Content-Type の値。
	contTypeJson = "application/json"
)
