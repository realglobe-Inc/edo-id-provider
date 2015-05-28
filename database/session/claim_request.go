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
type ClaimRequest struct {
	// id_token
	idTok map[string]*ClaimEntry
	// userinfo
	acnt map[string]*ClaimEntry
}

// ID トークンに入れて返すように要求されているクレームの情報を返す。
func (this *ClaimRequest) IdTokenEntries() map[string]*ClaimEntry {
	if this == nil { // あんまり良くないと思うが。
		return nil
	}
	return this.idTok
}

// アカウント情報エンドポイントから返すように要求されているクレームの情報を返す。
func (this *ClaimRequest) AccountEntries() map[string]*ClaimEntry {
	if this == nil { // あんまり良くないと思うが。
		return nil
	}
	return this.acnt
}

// クレーム名を返す。
// clms: 必須クレーム名。
// optClms: 必須でないクレーム名。
func (this *ClaimRequest) Names() (clms, optClms map[string]bool) {
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
func (this *ClaimRequest) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id_token": Claims(this.idTok),
		"userinfo": Claims(this.acnt),
	})
}

func (this *ClaimRequest) UnmarshalJSON(data []byte) error {
	var buff struct {
		Acnt  Claims `json:"userinfo"`
		IdTok Claims `json:"id_token"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}
	this.acnt = buff.Acnt
	this.idTok = buff.IdTok
	return nil
}
