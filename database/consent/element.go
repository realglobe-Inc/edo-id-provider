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

import (
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2/bson"
)

// アカウントがどの TA にどの情報の提供を許可しているかという情報。
type Element struct {
	acnt string
	ta   string
	// 許可スコープ。
	scop Consent
	// 許可属性。
	attr Consent
}

func New(acnt, ta string) *Element {
	return &Element{
		acnt: acnt,
		ta:   ta,
		scop: map[string]bool{},
		attr: map[string]bool{},
	}
}

func (this *Element) copy() *Element {
	elem := New(this.acnt, this.ta)
	for k := range this.scop {
		elem.scop[k] = true
	}
	for k := range this.attr {
		elem.attr[k] = true
	}
	return elem
}

// アカウント ID を返す。
func (this *Element) Account() string {
	return this.acnt
}

// 許可される TA の ID を返す。
func (this *Element) Ta() string {
	return this.ta
}

// スコープの許可情報を返す。
func (this *Element) Scope() Consent {
	return this.scop
}

// 属性の許可情報を返す。
func (this *Element) Attribute() Consent {
	return this.attr
}

//  {
//      "account": <アカウント ID>,
//      "ta": <TA の ID>,
//      "scopes": <許可スコープ>,
//      "attributes": <許可属性>
//  }
func (this *Element) GetBSON() (interface{}, error) {
	if this == nil {
		return nil, nil
	}

	return map[string]interface{}{
		"account":    this.acnt,
		"ta":         this.ta,
		"scopes":     this.scop,
		"attributes": this.attr,
	}, nil
}

func (this *Element) SetBSON(raw bson.Raw) error {
	var buff struct {
		Acnt string  `bson:"account"`
		Ta   string  `bson:"ta"`
		Scop Consent `bson:"scopes"`
		Attr Consent `bson:"attributes"`
	}
	if err := raw.Unmarshal(&buff); err != nil {
		return erro.Wrap(err)
	}

	this.acnt = buff.Acnt
	this.ta = buff.Ta
	this.scop = buff.Scop
	this.attr = buff.Attr
	return nil
}
