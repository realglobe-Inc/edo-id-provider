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
	"container/list"
	"time"
)

// セッション。
type Element struct {
	id string
	// 有効期限。
	exp time.Time
	// 最後に選択されたアカウント。
	acnt *Account
	// 現在のリクエスト内容。
	req *Request
	// 現在発行されているチケット。
	tic string
	// 過去に選択されたアカウント。
	pastAcnts list.List
	// 最後に選択された表示言語。
	lang string
}

func New(id string, exp time.Time) *Element {
	return &Element{
		id:  id,
		exp: exp,
	}
}

// 履歴を引き継いだセッションを作成する。
func (this *Element) New(id string, exp time.Time) *Element {
	elem := &Element{
		id:   id,
		exp:  exp,
		lang: this.lang,
	}
	for e := this.pastAcnts.Back(); e != nil; e = e.Prev() {
		elem.pastAcnts.PushFront(e.Value.(*Account).New())
	}
	if this.acnt != nil {
		elem.addPastAccount(this.acnt.New(), MaxHistory)
	}
	return elem
}

// 過去に選択されたアカウントをいくつまで記憶するか。
// 最後に選択されたアカウントも含む。
var MaxHistory = 5

// ID を返す。
func (this *Element) Id() string {
	return this.id
}

// 有効期限を返す。
func (this *Element) ExpiresIn() time.Time {
	return this.exp
}

// 最後に選択されたアカウントを返す。
func (this *Element) Account() *Account {
	return this.acnt
}

// アカウントが選択されたことを反映させる。
func (this *Element) SelectAccount(acnt *Account) {
	if this.acnt == nil || this.acnt.Id() != acnt.Id() {
		this.removePastAccount(acnt)
		if this.acnt != nil {
			this.addPastAccount(this.acnt, MaxHistory-1)
		}
	}
	this.acnt = acnt
}

func (this *Element) addPastAccount(acnt *Account, max int) {
	for this.pastAcnts.Len() >= max {
		this.pastAcnts.Remove(this.pastAcnts.Back())
	}
	this.pastAcnts.PushFront(acnt)
	return
}

func (this *Element) removePastAccount(acnt *Account) {
	for elem := this.pastAcnts.Front(); elem != nil; elem = elem.Next() {
		if elem.Value.(*Account).Id() == acnt.Id() {
			this.pastAcnts.Remove(elem)
			return
		}
	}
}

// 現在のリクエスト内容を返す。
func (this *Element) Request() *Request {
	return this.req
}

// リクエスト内容を保存する。
func (this *Element) SetRequest(req *Request) {
	this.req = req
}

// 現在発行されているチケットを返す。
func (this *Element) Ticket() string {
	return this.tic
}

// チケットを保存する。
func (this *Element) SetTicket(tic string) {
	this.tic = tic
}

// 過去に選択されたアカウントを返す。
func (this *Element) SelectedAccounts() []*Account {
	a := []*Account{}
	if this.acnt != nil {
		a = append(a, this.acnt)
	}
	for elem := this.pastAcnts.Front(); elem != nil; elem = elem.Next() {
		a = append(a, elem.Value.(*Account))
	}
	return a
}

// 最後に選択された表示言語を返す。
func (this *Element) Language() string {
	return this.lang
}

// 表示言語を保存する。
func (this *Element) SetLanguage(lang string) {
	this.lang = lang
}

// 一時データを消す。
func (this *Element) Clear() {
	this.req = nil
	this.tic = ""
}
