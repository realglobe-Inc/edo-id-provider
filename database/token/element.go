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

package token

import (
	"encoding/json"
	"time"

	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
)

// アクセストークン情報。
type Element struct {
	id string
	// 無効にされたかどうか。
	inv bool
	// 有効期限。
	exp time.Time
	// アカウント ID。
	acnt string
	// 許可スコープ。
	scop map[string]bool
	// 許可属性。
	attrs map[string]bool
	// 要請元 TA の ID。
	ta string
	// 発行したアクセストークン。
	toks map[string]bool

	// 更新日時。
	date time.Time
}

func New(id string, exp time.Time, acnt string, scop, attrs map[string]bool, ta string) *Element {
	return newElement(id, false, exp, acnt, scop, attrs, ta, time.Now())
}
func newElement(id string, inv bool, exp time.Time, acnt string, scop, attrs map[string]bool, ta string, date time.Time) *Element {
	return &Element{
		inv:   inv,
		id:    id,
		exp:   exp,
		acnt:  acnt,
		scop:  scop,
		attrs: attrs,
		ta:    ta,
		date:  date,
	}
}

// ID を返す。
func (this *Element) Id() string {
	return this.id
}

// 無効にされているかどうか。
func (this *Element) Invalid() bool {
	return this.inv
}

// 無効にする。
func (this *Element) Invalidate() {
	this.inv = true
	this.date = time.Now()
}

// 有効期限を返す。
func (this *Element) Expires() time.Time {
	return this.exp
}

// アカウント ID を返す。
func (this *Element) Account() string {
	return this.acnt
}

// 許可スコープを返す。
func (this *Element) Scope() map[string]bool {
	return this.scop
}

// 許可属性を返す。
func (this *Element) Attributes() map[string]bool {
	return this.attrs
}

// TA の ID を返す。
func (this *Element) Ta() string {
	return this.ta
}

// 発行したアクセストークンを返す。
func (this *Element) Tokens() map[string]bool {
	return this.toks
}

// 発行したアクセストークンを保存する。
func (this *Element) AddToken(tok string) {
	if this.toks == nil {
		this.toks = map[string]bool{}
	}
	this.toks[tok] = true
	this.date = time.Now()
}

// 更新日時を返す。
func (this *Element) Date() time.Time {
	return this.date
}

//  {
//      "id": <ID>,
//      "invalid": <無効か>,
//      "expires": <有効期限>,
//      "account": <アカウント ID>,
//      "scope": [
//          <許可スコープ>,
//          ...
//      ],
//      "attributes": [
//          <許可属性>,
//          ...
//      ],
//      "client_id": <TA の ID>,
//      "tokens": [
//          <アクセストークン>,
//          ...
//      ],
//      "date": <更新日時>
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id":         this.id,
		"invalid":    this.inv,
		"expires":    this.exp,
		"account":    this.acnt,
		"scope":      strset.Set(this.scop),
		"attributes": strset.Set(this.attrs),
		"client_id":  this.ta,
		"tokens":     strset.Set(this.toks),
		"date":       this.date,
	})
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id    string     `json:"id"`
		Inv   bool       `json:"invalid"`
		Exp   time.Time  `json:"expires"`
		Acnt  string     `json:"account"`
		Scop  strset.Set `json:"scope"`
		Attrs strset.Set `json:"attributes"`
		Ta    string     `json:"client_id"`
		Toks  strset.Set `json:"tokens"`
		Date  time.Time  `json:"date"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.inv = buff.Inv
	this.exp = buff.Exp
	this.acnt = buff.Acnt
	this.scop = buff.Scop
	this.attrs = buff.Attrs
	this.ta = buff.Ta
	this.toks = buff.Toks
	this.date = buff.Date
	return nil
}
