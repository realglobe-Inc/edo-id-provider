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

package consent

import ()

// アカウントがどの TA にどの情報の提供を許可しているかという情報。
type Element struct {
	acntId string
	taId   string
	// 許可スコープ。
	scops map[string]bool
	// 許可属性。
	attrs map[string]bool
}

func New(acntId, taId string) *Element {
	return &Element{
		acntId: acntId,
		taId:   taId,
		scops:  map[string]bool{},
		attrs:  map[string]bool{},
	}
}

// アカウント ID を返す。
func (this *Element) AccountId() string {
	return this.acntId
}

// 許可される TA の ID を返す。
func (this *Element) TaId() string {
	return this.taId
}

// スコープが許可されているかどうか。
func (this *Element) ScopeAllowed(scop string) bool {
	return this.scops[scop]
}

// スコープが許可されたことを反映させる。
func (this *Element) AllowScope(scop string) {
	if this.scops == nil {
		this.scops = map[string]bool{}
	}
	this.scops[scop] = true
}

// スコープが拒否されたことを反映させる。
func (this *Element) DenyScope(scop string) {
	if this.scops == nil {
		this.scops = map[string]bool{}
	}
	delete(this.scops, scop)
}

// 属性が許可されているかどうか。
func (this *Element) AttributeAllowed(attr string) bool {
	return this.attrs[attr]
}

// 属性が許可されたことを反映させる。
func (this *Element) AllowAttribute(attr string) {
	if this.attrs == nil {
		this.attrs = map[string]bool{}
	}
	this.attrs[attr] = true
}

// 属性が拒否されたことを反映させる。
func (this *Element) DenyAttribute(attr string) {
	if this.attrs == nil {
		this.attrs = map[string]bool{}
	}
	delete(this.attrs, attr)
}