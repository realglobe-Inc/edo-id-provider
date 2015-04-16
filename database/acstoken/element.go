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

package acstoken

import (
	"time"
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
	return &Element{
		id:    id,
		exp:   exp,
		acnt:  acnt,
		scop:  scop,
		attrs: attrs,
		ta:    ta,
		date:  time.Now(),
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
func (this *Element) ExpiresIn() time.Time {
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
