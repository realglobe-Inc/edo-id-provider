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
	// 発行されるアクセストークンの有効期限。
	tokExp time.Time
	// 関連アカウント。
	relAcnts []*Account
	// 要請元 TA の ID。
	taFrom string
	// 要請先 TA の ID。
	taTo string
	// 発行したアクセストークン。
	tok string

	// 更新日時。
	date time.Time
}

func New(id string, exp time.Time, acnt *Account, scop map[string]bool, tokExp time.Time,
	relAcnts []*Account, taFrom, taTo string) *Element {
	return &Element{
		id:       id,
		exp:      exp,
		acnt:     acnt,
		scop:     scop,
		tokExp:   tokExp,
		relAcnts: relAcnts,
		taFrom:   taFrom,
		taTo:     taTo,
		date:     time.Now(),
	}
}

// ID を返す。
func (this *Element) Id() string {
	return this.id
}

// 有効期限を返す。
func (this *Element) ExpiresIn() time.Time {
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

// 発行されるアクセストークンの有効期限を返す。
func (this *Element) TokenExpiresIn() time.Time {
	return this.tokExp
}

// 関連アカウントを返す。
func (this *Element) RelatedAccounts() []*Account {
	return this.relAcnts
}

// 要請元 TA の ID を返す。
func (this *Element) TaFrom() string {
	return this.taFrom
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
