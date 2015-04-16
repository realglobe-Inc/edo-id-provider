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

import ()

// バックエンドのデータを他のプログラムと共有する前提。

// アカウントがどの TA にどの属性の提供を許可しているかという情報の格納庫。
type Db interface {
	// 取得。
	Get(acntId, taId string) (Element, error)

	// 保存。
	Save(elem *Element) error
}
