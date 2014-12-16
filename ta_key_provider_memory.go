package main

import (
	"crypto/rsa"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type MemoryTaKeyProvider struct {
	base driver.KeyValueStore
}

// スレッドセーフ。
func NewMemoryTaKeyProvider(staleDur, expiDur time.Duration) *MemoryTaKeyProvider {
	return &MemoryTaKeyProvider{driver.NewMemoryListedKeyValueStore(staleDur, expiDur)}
}

func (reg *MemoryTaKeyProvider) ServiceKey(servUuid string, caStmp *driver.Stamp) (servKey *rsa.PublicKey, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := reg.base.Get(servUuid, caStmp)
	if value != nil {
		servKey = value.(*rsa.PublicKey)
	}
	return servKey, newCaStmp, err
}

func (reg *MemoryTaKeyProvider) AddServiceKey(servUuid string, servKey *rsa.PublicKey) {
	reg.base.Put(servUuid, servKey)
}

func (reg *MemoryTaKeyProvider) RemoveServiceKey(servUuid string) {
	reg.base.Remove(servUuid)
}
