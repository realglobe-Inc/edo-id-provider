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

package main

import ()

// サポートするスコープと紐付くクレーム。
var knownScops = map[string]map[string]bool{
	// ID トークンの被発行権。
	"openid": {},
	// リフレッシュトークンの被発行権。
	//"offline_access": {},
	// 以下、クレーム集合の取得権。
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

// 知らないスコープを除く。
// 返り値は scops。
func stripUnknownScopes(scops map[string]bool) map[string]bool {
	for scop := range scops {
		if knownScops[scop] == nil {
			log.Debug("Remove " + scop)
			delete(scops, scop)
		}
	}
	return scops
}

// スコープに対応するクレームを返す。
// 返り値は自由に書き換えて良い。
func scopesToClaims(scops map[string]bool) map[string]bool {
	clms := map[string]bool{}
	for scop, ok := range scops {
		if !ok {
			continue
		}
		for clm, ok := range knownScops[scop] {
			if !ok {
				continue
			}
			clms[clm] = true
		}
	}
	return clms
}
