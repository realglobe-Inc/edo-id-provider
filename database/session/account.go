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
	"time"
)

// セッションに付属させるアカウント情報。
type Account struct {
	id string
	// ログイン名。
	name string
	// ログイン日時。
	loginDate time.Time
}

func NewAccount(id, name string) *Account {
	return &Account{
		id:   id,
		name: name,
	}
}

// 設定を引き継いだアカウント情報を作成する。
func (this *Account) New() *Account {
	return &Account{
		id:   this.id,
		name: this.name,
	}
}

// ID を返す。
func (this *Account) Id() string {
	return this.id
}

// ログイン名を返す。
func (this *Account) Name() string {
	return this.name
}

// ログインしているかどうか。
func (this *Account) LoggedIn() bool {
	return !this.loginDate.IsZero()
}

// ログイン日時を返す。
func (this *Account) LoginDate() time.Time {
	return this.loginDate
}

// ログインしたことを反映させる。
func (this *Account) Login() {
	this.loginDate = time.Now()
}
