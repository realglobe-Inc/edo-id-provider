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

package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type account struct {
	m map[string]interface{}

	// IdP 内で一意かつ変更されることのない ID。
	accId string
	// IdP 内で一意のログイン ID。
	accName string
	// パスワード。
	passwd string
	// 更新日時。
	upd time.Time
}

func newAccount(m map[string]interface{}) *account {
	a := &account{m: map[string]interface{}{}}
	for k, v := range m {
		a.m[k] = v
	}
	return a
}

func (this *account) id() string {
	if this.accId == "" {
		this.accId, _ = this.m["id"].(string)
	}
	return this.accId
}

func (this *account) name() string {
	if this.accName == "" {
		this.accName, _ = this.m["username"].(string)
	}
	return this.accName
}

func (this *account) password() string {
	if this.passwd == "" {
		this.passwd, _ = this.m["password"].(string)
	}
	return this.passwd
}

func (this *account) updateDate() time.Time {
	if this.upd.IsZero() {
		this.upd, _ = this.m["update_at"].(time.Time)
	}
	return this.upd
}

func (this *account) attribute(attrName string) interface{} {
	return this.m[attrName]
}

func (this *account) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.m)
}

func (this *account) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &this.m)
}

func (this *account) GetBSON() (interface{}, error) {
	return this.m, nil
}

func (this *account) SetBSON(raw bson.Raw) error {
	return raw.Unmarshal(&this.m)
}
