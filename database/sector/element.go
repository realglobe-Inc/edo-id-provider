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

package sector

import (
	"encoding/base64"

	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2/bson"
)

// セクタ固有のアカウント ID の計算に使う情報。
type Element struct {
	// セクタ ID。
	id string
	// ソルト値。
	salt []byte
}

func New(id string, salt []byte) *Element {
	return &Element{
		id:   id,
		salt: salt,
	}
}

// セクタ ID を返す。
func (this *Element) Id() string {
	return this.id
}

// ソルト値を返す。
func (this *Element) Salt() []byte {
	return this.salt
}

//  {
//      "id": <セクタ ID>,
//      "salt": <ソルト値>
//  }
func (this *Element) GetBSON() (interface{}, error) {
	if this == nil {
		return nil, nil
	}

	return map[string]interface{}{
		"id":   this.id,
		"salt": base64.RawURLEncoding.EncodeToString(this.salt),
	}, nil
}

func (this *Element) SetBSON(raw bson.Raw) error {
	var buff struct {
		Id   string `bson:"id"`
		Salt string `bson:"salt"`
	}
	if err := raw.Unmarshal(&buff); err != nil {
		return erro.Wrap(err)
	}

	salt, err := base64.RawURLEncoding.DecodeString(buff.Salt)
	if err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.salt = salt
	return nil
}
