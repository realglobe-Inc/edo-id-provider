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

package jti

// バックエンドのデータもこのプログラム専用の前提。

// JWT の ID の格納庫。
type Db interface {
	// 保存。
	// 発行者と ID が既存の有効期限が切れていないものと重複する場合は失敗する。
	SaveIfAbsent(elem *Element) (ok bool, err error)
}
