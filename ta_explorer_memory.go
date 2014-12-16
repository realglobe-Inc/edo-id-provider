package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type MemoryTaExplorer struct {
	base driver.KeyValueStore
}

// スレッドセーフ。
func NewMemoryTaExplorer(staleDur, expiDur time.Duration) *MemoryTaExplorer {
	return &MemoryTaExplorer{driver.NewMemoryListedKeyValueStore(staleDur, expiDur)}
}

func (reg *MemoryTaExplorer) ServiceUuid(servUri string, caStmp *driver.Stamp) (servUuid string, newCaStmp *driver.Stamp, err error) {
	return (&taExplorer{reg.base}).ServiceUuid(servUri, caStmp)
}

func (reg *MemoryTaExplorer) SetServiceUuids(uriToUuid map[string]string) {
	tree := newTaExplorerTree()
	tree.fromContainer(uriToUuid)
	reg.base.Put("list", tree)
}
