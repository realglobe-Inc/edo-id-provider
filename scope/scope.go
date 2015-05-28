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

// スコープ周りの関数
package scope

import ()

// サポートするスコープと紐付く属性。
var knownScops = map[string]map[string]bool{
	// ID トークンの被発行権。
	tagOpenid: nil,
	// リフレッシュトークンの被発行権。
	tagOffline_access: nil,
	// 以下、属性集合。
	"profile": {
		"name":               true,
		"family_name":        true,
		"given_name":         true,
		"middle_name":        true,
		"nickname":           true,
		"preferred_username": true,
		"profile":            true,
		"picture":            true,
		"website":            true,
		"gender":             true,
		"birthdate":          true,
		"zoneinfo":           true,
		"locale":             true,
		"updated_at":         true,
	},
	"email": {
		"email":          true,
		"email_verified": true,
	},
	"address": {
		"address": true,
	},
	"phone": {
		"phone_number":          true,
		"phone_number_verified": true,
	},
}

// スコープに紐付く属性からスコープへのマップ。
var attrToScop = func() map[string]string {
	m := map[string]string{}
	for scop, attrSet := range knownScops {
		for attr := range attrSet {
			m[attr] = scop
		}
	}
	return m
}()

// 許可必須スコープ。
var essScops = map[string]bool{
	tagOpenid:         true,
	tagOffline_access: true,
}

// 知らないスコープを除く。
func RemoveUnknown(scops map[string]bool) map[string]bool {
	res := map[string]bool{}
	for scop := range scops {
		if _, ok := knownScops[scop]; !ok {
			log.Warn("Remove unknown scope " + scop)
			continue
		}
		res[scop] = true
	}
	return res
}

// スコープに対応する属性を返す。
func Attributes(scop string) map[string]bool {
	// 防御的コピー。
	attrs := map[string]bool{}
	for attr := range knownScops[scop] {
		attrs[attr] = true
	}
	return attrs
}

// 属性に対応するスコープを返す。
func FromAttribute(attr string) string {
	return attrToScop[attr]
}

// 許可が必須かどうか。
func IsEssential(scop string) bool {
	return essScops[scop]
}
