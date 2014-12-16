package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type MemoryUserNameIndex struct {
	base driver.KeyValueStore
}

// スレッドセーフ。
func NewMemoryUserNameIndex(staleDur, expiDur time.Duration) *MemoryUserNameIndex {
	return &MemoryUserNameIndex{driver.NewMemoryListedKeyValueStore(staleDur, expiDur)}
}

func (reg *MemoryUserNameIndex) UserUuid(usrName string, caStmp *driver.Stamp) (usrUuid string, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := reg.base.Get(usrName, caStmp)
	if err != nil {
		return "", nil, erro.Wrap(err)
	} else if value == nil || value == "" {
		return "", newCaStmp, nil
	}
	return value.(string), newCaStmp, err
}

func (reg *MemoryUserNameIndex) AddUserUuid(usrName, usrUuid string) {
	reg.base.Put(usrName, usrUuid)
}

func (reg *MemoryUserNameIndex) RemoveIdProvider(usrName string) {
	reg.base.Remove(usrName)
}
