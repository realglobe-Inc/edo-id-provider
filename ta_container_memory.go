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

package main

import (
	"github.com/realglobe-Inc/edo-lib/driver"
	"time"
)

type memoryTaContainer taContainerImpl

// スレッドセーフ。
func newMemoryTaContainer(staleDur, expiDur time.Duration) *memoryTaContainer {
	return (*memoryTaContainer)(&taContainerImpl{driver.NewMemoryListedKeyValueStore(staleDur, expiDur)})
}

func (this *memoryTaContainer) get(taId string) (*ta, error) {
	return ((*taContainerImpl)(this)).get(taId)
}

func (this *memoryTaContainer) close() error {
	return ((*taContainerImpl)(this)).close()
}

func (this *memoryTaContainer) add(ta *ta) {
	((*taContainerImpl)(this)).base.(driver.KeyValueStore).Put(ta.id(), ta)
}
