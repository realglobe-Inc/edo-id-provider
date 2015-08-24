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

package pairwise

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/realglobe-Inc/edo-lib/hash"
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2/bson"
)

// セクタ固有のアカウント ID の情報。
type Element struct {
	// 真のアカウント ID。
	acnt string
	// TA のセクタ ID。
	sect string
	// セクタ固有のアカウント ID。
	pw string
}

func New(acnt, sect, pw string) *Element {
	return &Element{
		acnt: acnt,
		sect: sect,
		pw:   pw,
	}
}

// 真のアカウント ID を返す。
func (this *Element) Account() string {
	return this.acnt
}

// TA のセクタ ID を返す。
func (this *Element) Sector() string {
	return this.sect
}

// セクタ固有のアカウント ID を返す。
func (this *Element) Pairwise() string {
	return this.pw
}

// セクタ固有のアカウントを計算する。
func Generate(acnt, sect string, salt []byte) *Element {
	h := hash.Hashing(sha256.New(), []byte(acnt), []byte{0}, []byte(sect), []byte{0}, salt)
	return New(acnt, sect, base64.RawURLEncoding.EncodeToString(h))
}

//  {
//      "account": <アカウント ID>,
//      "sector": <セクタ ID>,
//      "pairwise": <セクタ固有のアカウント ID>
//  }
func (this *Element) GetBSON() (interface{}, error) {
	if this == nil {
		return nil, nil
	}

	return map[string]interface{}{
		"account":  this.acnt,
		"sector":   this.sect,
		"pairwise": this.pw,
	}, nil
}

func (this *Element) SetBSON(raw bson.Raw) error {
	var buff struct {
		Acnt string `bson:"account"`
		Sect string `bson:"sector"`
		Pw   string `bson:"pairwise"`
	}
	if err := raw.Unmarshal(&buff); err != nil {
		return erro.Wrap(err)
	}

	this.acnt = buff.Acnt
	this.sect = buff.Sect
	this.pw = buff.Pw
	return nil
}
