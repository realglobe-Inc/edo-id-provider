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

package authcode

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

// 認可コード情報。
type Element struct {
	id string
	// 有効期限。
	exp time.Time
	// アカウント ID。
	acnt string
	// ログイン日時。
	lginDate time.Time
	// 許可スコープ。
	scop map[string]bool
	// ID トークンで提供可能な許可属性。
	tokAttrs map[string]bool
	// アカウント情報エンドポイントで提供可能な許可属性。
	acntAttrs map[string]bool
	// 要請元 TA の ID。
	ta string
	// 元になったリクエストの redirect_uri。
	rediUri string
	// 元になったリクエストの nonce。
	nonc string
	// 発行したアクセストークン。
	tok string

	// 更新日時。
	date time.Time
}

func New(id string, exp time.Time, acnt string, lginDate time.Time, scop, tokAttrs,
	acntAttrs map[string]bool, ta, rediUri, nonc string) *Element {
	return newElement(id, exp, acnt, lginDate, scop, tokAttrs, acntAttrs, ta, rediUri, nonc, time.Now())
}

func newElement(id string, exp time.Time, acnt string, lginDate time.Time, scop, tokAttrs,
	acntAttrs map[string]bool, ta, rediUri, nonc string, date time.Time) *Element {
	return &Element{
		id:        id,
		exp:       exp,
		acnt:      acnt,
		lginDate:  lginDate,
		scop:      scop,
		tokAttrs:  tokAttrs,
		acntAttrs: acntAttrs,
		ta:        ta,
		rediUri:   rediUri,
		nonc:      nonc,
		date:      date,
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

// アカウント ID を返す。
func (this *Element) Account() string {
	return this.acnt
}

// ログイン日時を返す。
func (this *Element) LoginDate() time.Time {
	return this.lginDate
}

// 許可スコープを返す。
func (this *Element) Scope() map[string]bool {
	return this.scop
}

// ID トークンでの提供可能属性を返す。
func (this *Element) IdTokenAttributes() map[string]bool {
	return this.tokAttrs
}

// アカウント情報エンドポイントでの提供可能属性を返す。
func (this *Element) AccountAttributes() map[string]bool {
	return this.acntAttrs
}

// TA の ID を返す。
func (this *Element) Ta() string {
	return this.ta
}

// 元になったリクエストの redirect_uri を返す。
func (this *Element) RedirectUri() string {
	return this.rediUri
}

// 元になったリクエストの nonce を返す。
func (this *Element) Nonce() string {
	return this.nonc
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
//      "account": <アカウント ID>,
//      "login_date": <ログイン日時>,
//      "scope": [
//          <許可スコープ>,
//          ...
//      ],
//      "id_token": [
//          <ID トークンでの許可属性>,
//          ...
//      ],
//      "userinfo": [
//          <アカウント情報エンドポイントでの許可属性>,
//      ],
//      "client_id": <TA の ID>,
//      "redirect_uri": <リダイレクトエンドポイント>,
//      "nonce": <nonce 値>,
//      "token": <アクセストークン>,
//      "date": <更新日時>
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id":           this.id,
		"expires":      this.exp,
		"account":      this.acnt,
		"login_date":   this.lginDate,
		"scope":        strset.Set(this.scop),
		"id_token":     strset.Set(this.tokAttrs),
		"userinfo":     strset.Set(this.acntAttrs),
		"client_id":    this.ta,
		"redirect_uri": this.rediUri,
		"nonce":        this.nonc,
		"token":        this.tok,
		"date":         this.date,
	})
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id        string     `json:"id"`
		Exp       time.Time  `json:"expires"`
		Acnt      string     `json:"account"`
		LginDate  time.Time  `json:"login_date"`
		Scop      strset.Set `json:"scope"`
		TokAttrs  strset.Set `json:"id_token"`
		AcntAttrs strset.Set `json:"userinfo"`
		Ta        string     `json:"client_id"`
		RediUri   string     `json:"redirect_uri"`
		Nonc      string     `json:"nonce"`
		Tok       string     `json:"token"`
		Date      time.Time  `json:"date"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.exp = buff.Exp
	this.acnt = buff.Acnt
	this.lginDate = buff.LginDate
	this.scop = buff.Scop
	this.tokAttrs = buff.TokAttrs
	this.acntAttrs = buff.AcntAttrs
	this.ta = buff.Ta
	this.rediUri = buff.RediUri
	this.nonc = buff.Nonc
	this.tok = buff.Tok
	this.date = buff.Date
	return nil
}
