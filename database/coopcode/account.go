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
	"encoding/json"
)

// 仲介コードに付属させるアカウント情報。
type Account struct {
	id string
	// アカウントタグ。
	tag string
}

func NewAccount(id, tag string) *Account {
	return &Account{
		id:  id,
		tag: tag,
	}
}

// ID を返す。
func (this *Account) Id() string {
	return this.id
}

// アカウントタグを返す。
func (this *Account) Tag() string {
	return this.tag
}

//  {
//      "id": <ID>,
//      "tag": <アカウントタグ>
//  }
func (this *Account) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]string{
		"id":  this.id,
		"tag": this.tag,
	})
}

func (this *Account) UnmarshalJSON(data []byte) error {
	var buff struct {
		Id  string `json:"id"`
		Tag string `json:"tag"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return err
	}
	this.id = buff.Id
	this.tag = buff.Tag
	return nil
}
