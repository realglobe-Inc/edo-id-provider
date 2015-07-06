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
	"reflect"

	"github.com/realglobe-Inc/edo-id-provider/claims"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/go-lib/erro"
)

// clms に応えられるかどうか。
// 返り値は応えられない場合のみ非 nil。
func CheckClaims(acnt account.Element, clms claims.Claims) error {
	if clms == nil {
		return nil
	}

	for clm, ent := range clms {
		attr := acnt.Attribute(clm)
		if ent.Essential() && attr == nil {
			return erro.New("account has no essential claim " + clm)
		} else if ent.Value() != nil && !reflect.DeepEqual(attr, ent.Value()) {
			return erro.New("account's claim "+clm+" is not ", ent.Value())
		} else if ent.Values() == nil {
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
			return erro.New("account's claim "+clm+" is not in ", ent.Values())
		}
	}
	return nil
}
