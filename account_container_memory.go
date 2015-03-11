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

type memoryAccountContainer accountContainerImpl

// スレッドセーフ。
func newMemoryAccountContainer(staleDur, expiDur time.Duration) *memoryAccountContainer {
	return (*memoryAccountContainer)(&accountContainerImpl{
		driver.NewMemoryListedKeyValueStore(staleDur, expiDur),
		driver.NewMemoryListedKeyValueStore(staleDur, expiDur),
	})
}

func (this *memoryAccountContainer) get(accId string) (*account, error) {
	return ((*accountContainerImpl)(this)).get(accId)
}

func (this *memoryAccountContainer) getByName(nameId string) (*account, error) {
	return ((*accountContainerImpl)(this)).getByName(nameId)
}

func (this *memoryAccountContainer) close() error {
	return ((*accountContainerImpl)(this)).close()
}

func (this *memoryAccountContainer) add(acc *account) {
	((*accountContainerImpl)(this)).idToAcc.(driver.KeyValueStore).Put(acc.id(), acc)
	((*accountContainerImpl)(this)).nameToAcc.(driver.KeyValueStore).Put(acc.name(), acc)
}
