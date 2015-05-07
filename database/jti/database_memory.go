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

import (
	"sync"
	"time"
)

// メモリ上の JWT の ID の格納庫。
type memoryDb struct {
	lock          sync.Mutex
	issToIdToElem map[string]map[string]*Element
}

func NewMemoryDb() Db {
	return &memoryDb{
		issToIdToElem: map[string]map[string]*Element{},
	}
}

// 保存。
func (this *memoryDb) SaveIfAbsent(elem *Element) (ok bool, err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	idToElem := this.issToIdToElem[elem.Issuer()]
	if idToElem == nil {
		idToElem = map[string]*Element{}
		this.issToIdToElem[elem.Issuer()] = idToElem
	} else if saved := idToElem[elem.Id()]; saved != nil || !time.Now().After(saved.Expires()) {
		return false, nil
	}

	idToElem[elem.Id()] = elem
	return true, nil
}
