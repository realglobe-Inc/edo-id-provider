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
	"github.com/realglobe-Inc/edo-lib/duration"
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
	// 許可スコープ。
	scop map[string]bool
	// 発行されるアクセストークンの有効期間。
	tokExpIn time.Duration
	// 関連アカウント。
	relAcnts []*Account
	// 要請元 TA の ID。
	taFr string
	// 要請先 TA の ID。
	taTo string
	// 発行したアクセストークン。
	tok string

	// 更新日時。
	date time.Time
}

func New(id string, exp time.Time, acnt *Account, scop map[string]bool, tokExpIn time.Duration,
	relAcnts []*Account, taFr, taTo string) *Element {
	return &Element{
		id:       id,
		exp:      exp,
		acnt:     acnt,
		scop:     scop,
		tokExpIn: tokExpIn,
		relAcnts: relAcnts,
		taFr:     taFr,
		taTo:     taTo,
		date:     time.Now(),
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

// 許可スコープを返す。
func (this *Element) Scope() map[string]bool {
	return this.scop
}

// 発行されるアクセストークンの有効期間を返す。
func (this *Element) TokenExpiresIn() time.Duration {
	return this.tokExpIn
}

// 関連アカウントを返す。
func (this *Element) RelatedAccounts() []*Account {
	return this.relAcnts
}

// 要請元 TA の ID を返す。
func (this *Element) TaFrom() string {
	return this.taFr
}

// 要請先 TA の ID を返す。
func (this *Element) TaTo() string {
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
//      "account": <主体>,
//      "scope": [
//          <許可スコープ>,
//          ...
//      ],
//      "token_expires_in": <発行されるアクセストークンの有効期間>,
//      "related_accounts": [
//          <主体でないアカウント>,
//          ...
//      ],
//      "ta_from": <要請元 TA の ID>,
//      "ta_to": <要請先 TA の ID>,
//      "token": <アクセストークン>,
//      "date": <更新日時>,
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id":               this.id,
		"expires":          this.exp,
		"account":          this.acnt,
		"scope":            strset.Set(this.scop),
		"token_expires_in": duration.Duration(this.tokExpIn),
		"related_accounts": this.relAcnts,
		"ta_from":          this.taFr,
		"ta_to":            this.taTo,
		"token":            this.tok,
		"date":             this.date,
	})
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id       string            `json:"id"`
		Exp      time.Time         `json:"expires"`
		Acnt     *Account          `json:"account"`
		Scop     strset.Set        `json:"scope"`
		TokExpIn duration.Duration `json:"token_expires_in"`
		RelAcnts []*Account        `json:"related_accounts"`
		TaFr     string            `json:"ta_from"`
		TaTo     string            `json:"ta_to"`
		Tok      string            `json:"token"`
		Date     time.Time         `json:"date"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.exp = buff.Exp
	this.acnt = buff.Acnt
	this.scop = buff.Scop
	this.tokExpIn = time.Duration(buff.TokExpIn)
	this.relAcnts = buff.RelAcnts
	this.taFr = buff.TaFr
	this.taTo = buff.TaTo
	this.tok = buff.Tok
	this.date = buff.Date
	return nil
}
