package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type MemoryUserAttributeRegistry struct {
	driver.KeyValueStore
}

// スレッドセーフ。
func NewMemoryUserAttributeRegistry(staleDur, expiDur time.Duration) *MemoryUserAttributeRegistry {
	return &MemoryUserAttributeRegistry{driver.NewMemoryListedKeyValueStore(staleDur, expiDur)}
}

func (reg *MemoryUserAttributeRegistry) UserAttribute(usrUuid, attrName string, caStmp *driver.Stamp) (usrAttr interface{}, newCaStmp *driver.Stamp, err error) {
	return reg.Get(usrUuid+"/"+attrName, caStmp)
}

func (reg *MemoryUserAttributeRegistry) AddUserAttribute(usrUuid, attrName string, usrAttr interface{}) {
	reg.Put(usrUuid+"/"+attrName, usrAttr)
}

func (reg *MemoryUserAttributeRegistry) RemoveIdProvider(usrUuid, attrName string) {
	reg.Remove(usrUuid + "/" + attrName)
}
