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

import ()

// メモリ上のアカウント情報の格納庫。
type memoryDb struct {
	idToElem   map[string]Element
	nameToElem map[string]Element
}

// 追加。
func (this *memoryDb) add(elem Element) {
	this.idToElem[elem.Id()] = elem
	this.nameToElem[elem.Name()] = elem
}

func NewMemoryDb(elems []Element) Db {
	db := &memoryDb{
		map[string]Element{},
		map[string]Element{},
	}
	for _, elem := range elems {
		db.add(elem)
	}
	return db
}

// 取得。
func (this *memoryDb) Get(id string) (Element, error) {
	return this.idToElem[id], nil
}

// ログイン名で取得。
func (this *memoryDb) GetByName(name string) (Element, error) {
	return this.nameToElem[name], nil
}
