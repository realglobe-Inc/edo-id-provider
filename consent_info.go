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

import (
	"github.com/realglobe-Inc/edo-id-provider/database/session"
)

// 許可情報。
// database/consent.Element が実装になるように。
type consentInfo interface {
	// スコープが許可されているかどうか。
	ScopeAllowed(scop string) bool
	// 属性が許可されているかどうか。
	AttributeAllowed(attr string) bool
}

type consentInfoImpl struct {
	scop  map[string]bool
	attrs map[string]bool
}

func newConsentInfo(scop, attrs map[string]bool) *consentInfoImpl {
	return &consentInfoImpl{scop, attrs}
}

func (this *consentInfoImpl) ScopeAllowed(scop string) bool {
	return this.scop[scop]
}

func (this *consentInfoImpl) AttributeAllowed(attr string) bool {
	return this.attrs[attr]
}

// ok: 同意 consInf で要求 reqScop, reqClms に応えられるかどうか。
// scop: 応えられるスコープ。
// tokAttrs: 応えられる場合に ID トークンに含ませる属性。
// acntAttrs: 応えられる場合にアカウント情報エンドポイントから返す属性。
func satisfiable(consInf consentInfo, reqScop map[string]bool, reqClm *session.Claim) (ok bool, scop, tokAttrs, acntAttrs map[string]bool) {

	scop = map[string]bool{}
	for s := range reqScop {
		if consInf.ScopeAllowed(s) {
			scop[s] = true
		} else if scopeEssential(s) {
			return false, nil, nil, nil
		}
	}

	tokAttrs = map[string]bool{}
	acntAttrs = scopeToClaims(scop) // スコープで許可された属性も加える。

	if reqClm != nil {
		for clm, ent := range reqClm.IdTokenEntries() {
			// スコープで許可された属性も許す。
			if consInf.AttributeAllowed(clm) || consInf.ScopeAllowed(attrToScop[clm]) {
				tokAttrs[clm] = true
				continue
			}
			if ent.Essential() {
				return false, nil, nil, nil
			}
		}

		for clm, ent := range reqClm.AccountEntries() {
			// スコープで許可された属性も許す。
			if acntAttrs[clm] {
				continue
			} else if consInf.AttributeAllowed(clm) || consInf.ScopeAllowed(attrToScop[clm]) {
				acntAttrs[clm] = true
				continue
			}
			if ent.Essential() {
				return false, nil, nil, nil
			}
		}
	}

	return true, scop, tokAttrs, acntAttrs
}
