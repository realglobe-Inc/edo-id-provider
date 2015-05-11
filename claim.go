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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/go-lib/erro"
	"reflect"
)

// 要求 reqClm に応えられるかどうか。
// err: 応えられない場合のみ非 nil。
func checkContradiction(acnt account.Element, reqClm *session.Claim) error {
	if reqClm == nil {
		return nil
	}

	for _, ents := range []map[string]*session.ClaimEntry{reqClm.AccountEntries(), reqClm.IdTokenEntries()} {
		for clmName, ent := range ents {
			if ent == nil {
				// Voluntary claim.
				continue
			}

			attr := acnt.Attribute(clmName)
			if ent.Essential() && attr == nil {
				return erro.New("no essential claim " + clmName)
			}
			if ent.Value() != nil && !reflect.DeepEqual(attr, ent.Value()) {
				return erro.New("claim "+clmName+" is not ", ent.Value())
			}
			if ent.Values() == nil {
				continue
			}
			// values
			ok := false
			for _, v := range ent.Values() {
				if reflect.DeepEqual(attr, v) {
					ok = true
					break
				}
			}
			if !ok {
				return erro.New("claim "+clmName+" is not in ", ent.Values())
			}
		}
	}
	return nil
}
