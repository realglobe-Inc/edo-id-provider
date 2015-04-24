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
	"time"
)

// バックエンドのデータもこのプログラム専用の前提。

// 仲介コード情報の格納庫。
type Db interface {
	// 取得。
	Get(id string) (*Element, error)

	// 保存。
	// exp: 保存期限。この期間以降は Get や Replace できなくて良い。
	Save(elem *Element, exp time.Time) error

	// 上書き。
	// savedDate が保存されている要素の更新日時と同じでなければ失敗する。
	Replace(elem *Element, savedDate time.Time) (ok bool, err error)
}
