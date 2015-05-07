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
)

// OpenID Connect Core 1.0 Section 5.5 を参照。

// 認証リクエストの claims パラメータ。
type Claim struct {
	// id_token
	idTok map[string]*ClaimEntry
	// userinfo
	acnt map[string]*ClaimEntry
}

// ID トークンに入れて返すように要求されているクレームの情報を返す。
func (this *Claim) IdTokenEntries() map[string]*ClaimEntry {
	return this.idTok
}

// アカウント情報エンドポイントから返すように要求されているクレームの情報を返す。
func (this *Claim) AccountEntries() map[string]*ClaimEntry {
	return this.acnt
}

// クレーム名を返す。
// clms: 必須クレーム名。
// optClms: 必須でないクレーム名。
func (this *Claim) Names() (clms, optClms map[string]bool) {
	clms = map[string]bool{}
	optClms = map[string]bool{}
	for _, set := range []map[string]*ClaimEntry{this.acnt, this.idTok} {
		for clm, ent := range set {
			if ent != nil && ent.Essential() {
				clms[clm] = true
				delete(optClms, clm)
			} else if !clms[clm] {
				optClms[clm] = true
			}
		}
	}
	return clms, optClms
}

//  {
//      "id_token": {
//          <属性名>: <ClaimEntry>,
//          ...
//      },
//      "userinfo": {
//          <属性名>: <ClaimEntry>,
//          ...
//      }
//  }
func (this *Claim) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id_token": this.idTok,
		"userinfo": this.acnt,
	})
}

func (this *Claim) UnmarshalJSON(data []byte) error {
	var buff struct {
		Acnt  map[string]*ClaimEntry `json:"userinfo"`
		IdTok map[string]*ClaimEntry `json:"id_token"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}
	this.acnt = buff.Acnt
	this.idTok = buff.IdTok
	return nil
}
