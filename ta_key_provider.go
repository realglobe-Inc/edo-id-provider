package main

import (
	"crypto/rsa"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

type TaKeyProvider interface {
	// サービスの公開鍵を返す。
	ServiceKey(servUuid string, caStmp *driver.Stamp) (servKey *rsa.PublicKey, newCaStmp *driver.Stamp, err error)
}

// 骨組み。
type taKeyProvider struct {
	base driver.KeyValueStore
}

func newTaKeyProvider(base driver.KeyValueStore) *taKeyProvider {
	return &taKeyProvider{base}
}

func (reg *taKeyProvider) ServiceKey(servUuid string, caStmp *driver.Stamp) (servKey *rsa.PublicKey, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := reg.base.Get(servUuid, caStmp)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if value == nil || value == "" {
		return nil, newCaStmp, nil
	}

	return value.(*rsa.PublicKey), newCaStmp, nil
}
