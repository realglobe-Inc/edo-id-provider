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
	"encoding/json"
	rist "github.com/realglobe-Inc/edo-lib/list"
	"github.com/realglobe-Inc/go-lib/erro"
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
	pastAcnts *list.List
	// 最後に選択された表示言語。
	lang string

	// 以下、作業用。

	// 読み込まれたセッションかどうか。
	saved bool
}

// 防御的コピー用。
func (this *Element) copy() *Element {
	elem := New(this.id, this.exp)
	if this.acnt != nil {
		elem.acnt = this.acnt.copy()
	}
	elem.req = this.req
	elem.tic = this.tic
	elem.tic = this.tic
	for e := this.pastAcnts.Back(); e != nil; e = e.Prev() {
		elem.pastAcnts.PushFront(e.Value.(*Account).copy())
	}
	elem.lang = this.lang
	return elem
}

func New(id string, exp time.Time) *Element {
	return &Element{
		id:        id,
		exp:       exp,
		pastAcnts: list.New(),
	}
}

// 履歴を引き継いだセッションを作成する。
func (this *Element) New(id string, exp time.Time) *Element {
	elem := &Element{
		id:        id,
		exp:       exp,
		lang:      this.lang,
		pastAcnts: list.New(),
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
func (this *Element) Expires() time.Time {
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

// 読み込まれたセッションかどうか。
func (this *Element) Saved() bool {
	return this.saved
}

func (this *Element) setSaved() {
	this.saved = true
}

//  {
//      "id": <ID>,
//      "expires": <有効期限>,
//      "account": <主アカウント>,
//      "request": <リクエスト内容>,
//      "ticket": <チケット>,
//      "past_accounts": [
//          <既ログインアカウント>,
//          ...
//      ],
//      "locale": <表示言語>
//  }
func (this *Element) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"id":            this.id,
		"expires":       this.exp,
		"account":       this.acnt,
		"request":       this.req,
		"ticket":        this.tic,
		"past_accounts": (*rist.List)(this.pastAcnts),
		"locale":        this.lang,
	})
}

func (this *Element) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id        string     `json:"id"`
		Exp       time.Time  `json:"expires"`
		Acnt      *Account   `json:"account"`
		Req       *Request   `json:"request"`
		Tic       string     `json:"ticket"`
		PastAcnts *rist.List `json:"past_accounts"`
		Lang      string     `json:"locale"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.exp = buff.Exp
	this.acnt = buff.Acnt
	this.req = buff.Req
	this.tic = buff.Tic
	this.pastAcnts = (*list.List)(buff.PastAcnts)
	if this.pastAcnts == nil {
		this.pastAcnts = list.New()
	}
	this.lang = buff.Lang
	return nil
}
