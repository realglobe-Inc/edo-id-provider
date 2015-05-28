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

package coopcode

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

// 仲介コード情報。
type Element struct {
	id string
	// 有効期限。
	exp time.Time
	// 主体。
	acnt *Account
	// 元になったアクセストークン。
	srcTok string
	// 許可スコープ。
	scop map[string]bool
	// 発行されるアクセストークンの有効期限。
	tokExp time.Time
	// 関連アカウント。
	acnts []*Account
	// 要請元 TA の ID。
	taFr string
	// 要請先 TA の ID。
	taTo string
	// 発行したアクセストークン。
	tok string

	// 更新日時。
	date time.Time
}

func New(id string, exp time.Time, acnt *Account, srcTok string, scop map[string]bool,
	tokExp time.Time, acnts []*Account, taFr, taTo string) *Element {
	return &Element{
		id:     id,
		exp:    exp,
		acnt:   acnt,
		srcTok: srcTok,
		scop:   scop,
		tokExp: tokExp,
		acnts:  acnts,
		taFr:   taFr,
		taTo:   taTo,
		date:   time.Now(),
	}
}

// ID を返す。
func (this *Element) Id() string {
	return this.id
}

// 有効期限を返す。
func (this *Element) Expires() time.Time {
	return this.exp
}

// アカウントを返す。
func (this *Element) Account() *Account {
	return this.acnt
}

// 元になったアクセストークンを返す。
func (this *Element) SourceToken() string {
	return this.srcTok
}

// 許可スコープを返す。
func (this *Element) Scope() map[string]bool {
	return this.scop
}

// 発行されるアクセストークンの有効期限を返す。
func (this *Element) TokenExpires() time.Time {
	return this.tokExp
}

// 関連アカウントを返す。
func (this *Element) Accounts() []*Account {
	return this.acnts
}

// 要請元 TA の ID を返す。
func (this *Element) FromTa() string {
	return this.taFr
}

// 要請先 TA の ID を返す。
func (this *Element) ToTa() string {
	return this.taTo
}

// 発行したアクセストークンを返す。
func (this *Element) Token() string {
	return this.tok
}

// 発行したアクセストークンを保存する。
func (this *Element) SetToken(tok string) {
	this.tok = tok
	this.date = time.Now()
}

// 更新日時を返す。
func (this *Element) Date() time.Time {
	return this.date
}

//  {
//      "id": <ID>,
//      "expires": <有効期限>,
//      "user": <主体>,
//      "source_token": <元になったアクセストークン>,
//      "scope": [
//          <発行されるアクセストークンの許可スコープ>,
//          ...
//      ],
//      "token_expires": <発行されるアクセストークンの有効期限>,
//      "users": [
//          <主体でないアカウント>,
//          ...
//      ],
//      "client_from": <要請元 TA の ID>,
//      "client_to": <要請先 TA の ID>,
//      "token": <アクセストークン>,
//      "date": <更新日時>,
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	m := map[string]interface{}{
		"id":          this.id,
		"expires":     this.exp,
		"user":        this.acnt,
		"client_from": this.taFr,
		"client_to":   this.taTo,
		"date":        this.date,
	}
	if this.srcTok != "" {
		m["source_token"] = this.srcTok
		m["scope"] = strset.Set(this.scop)
		m["token_expires"] = this.tokExp
	}
	if len(this.acnts) > 0 {
		m["users"] = this.acnts
	}
	if this.tok != "" {
		m["token"] = this.tok
	}
	return json.Marshal(m)
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id     string     `json:"id"`
		Exp    time.Time  `json:"expires"`
		Acnt   *Account   `json:"user"`
		SrcTok string     `json:"source_token"`
		Scop   strset.Set `json:"scope"`
		TokExp time.Time  `json:"token_expires"`
		Acnts  []*Account `json:"users"`
		TaFr   string     `json:"client_from"`
		ToTa   string     `json:"client_to"`
		Tok    string     `json:"token"`
		Date   time.Time  `json:"date"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.exp = buff.Exp
	this.acnt = buff.Acnt
	this.srcTok = buff.SrcTok
	this.scop = buff.Scop
	this.tokExp = buff.TokExp
	this.acnts = buff.Acnts
	this.taFr = buff.TaFr
	this.taTo = buff.ToTa
	this.tok = buff.Tok
	this.date = buff.Date
	return nil
}
