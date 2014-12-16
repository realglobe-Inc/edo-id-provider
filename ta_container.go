package main

import (
	"crypto/rsa"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

type ta struct {
	id   string
	name string
	// 登録された全ての redirect_uri。
	rediUris map[string]bool
	// kid から公開鍵へのマップ。
	pubKeys map[string]*rsa.PublicKey
}

type taContainer interface {
	get(taId string) (*ta, error)
}

type taContainerImpl struct {
	base driver.KeyValueStore
}

func (this *taContainerImpl) get(taId string) (*ta, error) {
	val, _, err := this.base.Get(taId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*ta), nil
}
