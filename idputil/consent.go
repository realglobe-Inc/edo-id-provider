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

import (
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/edo-id-provider/scope"
	"github.com/realglobe-Inc/go-lib/erro"
)

// 許可情報。
type Consent interface {
	// 許可されているか。
	Allow(name string) bool
}

// 提供するスコープを返す。
// 必須スコープが許可されない場合のみ err が非 nil。
func ProvidedScopes(scopCons Consent, reqScops map[string]bool) (scops map[string]bool, err error) {
	scops = map[string]bool{}
	for scop := range reqScops {
		if scopCons.Allow(scop) {
			scops[scop] = true
			continue
		}

		// 許可されなかった。

		if scope.IsEssential(scop) {
			return nil, erro.New("essential scope " + scop + " is not allowed")
		}
	}
	return scops, nil
}

// 提供する属性を返す。
// scops: 提供するスコープ。
// 必須属性が許可されない場合のみ err が非 nil。
func ProvidedAttributes(scopCons, attrCons Consent, scops map[string]bool, reqClms session.Claims) (attrs map[string]bool, err error) {
	attrs = map[string]bool{}

	// アカウント ID は必須。
	attrs[tagSub] = true

	// スコープで許可された属性を加える。
	for scop := range scops {
		for attr := range scope.Attributes(scop) {
			attrs[attr] = true
		}
	}

	for attr, ent := range reqClms {
		// 許可スコープに含まれる属性も許す。
		if attrCons.Allow(attr) || scopCons.Allow(scope.FromAttribute(attr)) {
			attrs[attr] = true
			continue
		}

		// 許可されなかった。

		if ent.Essential() {
			return nil, erro.New("essential attribute " + attr + " is not allowed")
		}
	}

	return attrs, nil
}
