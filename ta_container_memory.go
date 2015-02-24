package main

import (
	"github.com/realglobe-Inc/edo-toolkit/driver"
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
