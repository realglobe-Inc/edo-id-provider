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

import (
	"sync"
)

// メモリ上のアカウント情報の格納庫。
type memoryDb struct {
	lock           sync.Mutex
	acntToTaToElem map[string]map[string]*Element
}

func NewMemoryDb() Db {
	return &memoryDb{
		acntToTaToElem: map[string]map[string]*Element{},
	}
}

// 取得。
func (this *memoryDb) Get(acnt, ta string) (*Element, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	taToElem := this.acntToTaToElem[acnt]
	if taToElem == nil {
		return nil, nil
	}
	elem := taToElem[ta]
	if elem == nil {
		return nil, nil
	}
	// 防御的コピー。
	return elem.copy(), nil
}

// 保存。
func (this *memoryDb) Save(elem *Element) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	taToElem := this.acntToTaToElem[elem.Account()]
	if taToElem == nil {
		taToElem = map[string]*Element{}
		this.acntToTaToElem[elem.Account()] = taToElem
	}
	// 防御的コピー。
	taToElem[elem.Ta()] = elem.copy()
	return nil
}
