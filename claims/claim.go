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

// claims パラメータ関係。
package claims

import (
	"encoding/json"

	"github.com/realglobe-Inc/go-lib/erro"
)

// OpenID Connect Core 1.0 Section 5.5 を参照。

// 認証リクエストの claims.id_token や claims.userinfo パラメータの要素。
type Claim struct {
	// essential
	ess bool
	// value
	val interface{}
	// values
	vals []interface{}

	// 言語タグ。
	lang string
}

// 主にテスト用。
func New(ess bool, val interface{}, vals []interface{}, lang string) *Claim {
	return &Claim{
		ess:  ess,
		val:  val,
		vals: vals,
		lang: lang,
	}
}

func (this *Claim) Essential() bool {
	return this.ess
}

func (this *Claim) Value() interface{} {
	return this.val
}

func (this *Claim) Values() []interface{} {
	return this.vals
}

func (this *Claim) Language() string {
	return this.lang
}

func (this *Claim) setLanguage(lang string) {
	this.lang = lang
}

//  {
//      "essential": <必須か>,
//      "value": <指定値>,
//      "values": [
//          <候補値>,
//          ....
//      ]
//  }
func (this *Claim) MarshalJSON() (data []byte, err error) {
	var m map[string]interface{}
	if this.ess {
		if m == nil {
			m = map[string]interface{}{}
		}
		m["essential"] = true
	}
	if this.val != nil {
		if m == nil {
			m = map[string]interface{}{}
		}
		m["value"] = this.val
	}
	if this.vals != nil {
		if m == nil {
			m = map[string]interface{}{}
		}
		m["values"] = this.vals
	}
	return json.Marshal(m)
}

func (this *Claim) UnmarshalJSON(data []byte) error {
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
