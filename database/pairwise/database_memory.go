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
	lock         sync.Mutex
	taToPwToElem map[string]map[string]*Element
}

func NewMemoryDb() Db {
	return &memoryDb{
		taToPwToElem: map[string]map[string]*Element{},
	}
}

// TA 固有のアカウント ID による取得。
func (this *memoryDb) GetByPairwise(ta, pwAcnt string) (*Element, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	pwToElem := this.taToPwToElem[ta]
	if pwToElem == nil {
		return nil, nil
	}
	return pwToElem[pwAcnt], nil
}

// 保存。
func (this *memoryDb) Save(elem *Element) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	pwToElem := this.taToPwToElem[elem.Ta()]
	if pwToElem == nil {
		pwToElem = map[string]*Element{}
		this.taToPwToElem[elem.Ta()] = pwToElem
	}
	pwToElem[elem.PairwiseAccount()] = elem
	return nil
}
