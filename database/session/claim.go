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
	// userinfo
	acntInf map[string]*ClaimEntry
	// id_token
	idTok map[string]*ClaimEntry
}

// アカウント情報エンドポイントから返すように要求されているクレームの情報を返す。
func (this *Claim) WithAccountInfo() map[string]*ClaimEntry {
	return this.acntInf
}

// ID トークンに入れて返すように要求されているクレームの情報を返す。
func (this *Claim) WithIdToken() map[string]*ClaimEntry {
	return this.idTok
}

func (this *Claim) UnmarshalJSON(data []byte) error {
	var buff struct {
		AcntInf map[string]*ClaimEntry `json:"userinfo"`
		IdTok   map[string]*ClaimEntry `json:"id_token"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}
	this.acntInf = buff.AcntInf
	this.idTok = buff.IdTok
	return nil
}

type ClaimEntry struct {
	// essential
	ess bool
	// value
	val interface{}
	// values
	vals []interface{}
}

func (this *ClaimEntry) Essential() bool {
	return this.ess
}

func (this *ClaimEntry) Value() interface{} {
	return this.val
}

func (this *ClaimEntry) Values() []interface{} {
	return this.vals
}

func (this *ClaimEntry) UnmarshalJSON(data []byte) error {
	var buff struct {
		Ess  bool          `json:"essential"`
		Val  interface{}   `json:"value"`
		Vals []interface{} `json:"values"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}
	this.ess = buff.Ess
	this.val = buff.Val
	this.vals = buff.Vals
	return nil
}
