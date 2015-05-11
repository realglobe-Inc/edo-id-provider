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

package account

import (
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2/bson"
)

// アカウント情報の実装。
type element struct {
	id    string
	name  string
	auth  Authenticator
	attrs map[string]interface{}
}

// 主にテスト用。
func New(id, name string, auth Authenticator, attrs map[string]interface{}) Element {
	return newElement(id, name, auth, attrs)
}

func newElement(id, name string, auth Authenticator, attrs map[string]interface{}) *element {
	return &element{id, name, auth, attrs}
}

func (this *element) Id() string {
	return this.id
}

func (this *element) Name() string {
	return this.name
}

func (this *element) Authenticator() Authenticator {
	return this.auth
}

func (this *element) Attribute(attrName string) interface{} {
	return this.attrs[attrName]
}

func (this *element) SetAttribute(attrName string, attr interface{}) {
	this.attrs[attrName] = attr
}

//  {
//      "id": <ID>,
//      "username": <ログイン名>,
//      "authenticator": <認証器>,
//      <属性名>: <属性値>,
//      ...
//  }
func (this *element) SetBSON(raw bson.Raw) error {
	var buff struct {
		Id    string                 `bson:"id"`
		Name  string                 `bson:"username"`
		Auth  map[string]interface{} `bson:"authenticator"`
		Attrs map[string]interface{} `bson:",inline"`
	}
	if err := raw.Unmarshal(&buff); err != nil {
		return erro.Wrap(err)
	}

	auth, err := authenticatorFromMap(buff.Auth)
	if err != nil {
		return erro.Wrap(err)
	}

	this.id = buff.Id
	this.name = buff.Name
	this.auth = auth
	this.attrs = buff.Attrs
	return nil
}
