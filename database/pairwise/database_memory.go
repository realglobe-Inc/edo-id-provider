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

package pairwise

import (
	"sync"
)

// メモリ上の認可コード情報の格納庫。
type memoryDb struct {
	lock           sync.Mutex
	sectToPwToElem map[string]map[string]*Element
}

func NewMemoryDb() Db {
	return &memoryDb{
		sectToPwToElem: map[string]map[string]*Element{},
	}
}

// セクタ固有のアカウント ID による取得。
func (this *memoryDb) GetByPairwise(sect, pw string) (*Element, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	pwToElem := this.sectToPwToElem[sect]
	if pwToElem == nil {
		return nil, nil
	}
	return pwToElem[pw], nil
}

// 保存。
func (this *memoryDb) Save(elem *Element) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	pwToElem := this.sectToPwToElem[elem.Sector()]
	if pwToElem == nil {
		pwToElem = map[string]*Element{}
		this.sectToPwToElem[elem.Sector()] = pwToElem
	}
	pwToElem[elem.Pairwise()] = elem
	return nil
}
