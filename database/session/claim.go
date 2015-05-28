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

package session

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib/erro"
	"strings"
)

// 認証リクエストの claims パラメータの id_token や userinfo 要素。
type Claims map[string]*ClaimEntry

func (this Claims) MarshalJSON() (data []byte, err error) {
	if this == nil {
		return json.Marshal(nil)
	}
	m := map[string]*ClaimEntry{}
	for clm, ent := range this {
		if ent.Language() != "" {
			clm += "#" + ent.Language()
		}
		m[clm] = ent
	}
	return json.Marshal(m)
}

func (this *Claims) UnmarshalJSON(data []byte) error {
	m := map[string]*ClaimEntry{}
	if err := json.Unmarshal(data, &m); err != nil {
		return erro.Wrap(err)
	} else if m == nil {
		return nil
	}
	m2 := map[string]*ClaimEntry{}
	for clm, ent := range m {
		if ent == nil {
			ent = &ClaimEntry{}
		}
		if idx := strings.IndexRune(clm, '#'); idx >= 0 {
			ent.setLanguage(clm[idx+1:])
			clm = clm[:idx]
		}
		m2[clm] = ent
	}
	*this = m2
	return nil
}
