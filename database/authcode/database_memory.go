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

package authcode

import (
	"sync"
	"time"
)

// メモリ上の認可コード情報の格納庫。
type memoryDb struct {
	lock     sync.Mutex
	idToElem map[string]*Element
	idToExp  map[string]time.Time
}

func NewMemoryDb() Db {
	return &memoryDb{
		idToElem: map[string]*Element{},
		idToExp:  map[string]time.Time{},
	}
}

// 取得。
func (this *memoryDb) Get(id string) (*Element, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	elem := this.idToElem[id]
	if elem == nil {
		return nil, nil
	} else if time.Now().After(this.idToExp[id]) {
		delete(this.idToElem, id)
		delete(this.idToExp, id)
		return nil, nil
	}

	// Replace で使う更新日時が変わらないように防御的コピー。
	e := *elem
	return &e, nil
}

// 保存。
func (this *memoryDb) Save(elem *Element, exp time.Time) error {
	// Replace で使う更新日時が変わらないように防御的コピー。
	e := *elem

	this.lock.Lock()
	defer this.lock.Unlock()

	this.idToElem[elem.Id()] = &e
	this.idToExp[elem.Id()] = exp
	return nil
}

// 上書き。
func (this *memoryDb) Replace(elem *Element, savedDate time.Time) (ok bool, err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if saved := this.idToElem[elem.Id()]; saved == nil || !savedDate.Equal(saved.Date()) {
		return false, nil
	}

	// Replace で使う更新日時が変わらないように防御的コピー。
	e := *elem
	this.idToElem[elem.Id()] = &e
	return true, nil
}
