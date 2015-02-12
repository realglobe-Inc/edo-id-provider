package main

import (
	"github.com/realglobe-Inc/edo/driver"
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
