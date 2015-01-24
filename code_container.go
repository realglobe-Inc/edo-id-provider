package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type codeContainer interface {
	newId() (id string, err error)
	get(codId string) (*code, error)
	put(cod *code) error
}

type codeContainerImpl struct {
	base driver.TimeLimitedKeyValueStore

	idGenerator
	// 有効期限が切れてからも保持する期間。
	expiDur time.Duration
}

func (this *codeContainerImpl) put(cod *code) error {
	if _, err := this.base.Put(cod.id(), cod, cod.expirationDate()); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *codeContainerImpl) get(codId string) (*code, error) {
	val, _, err := this.base.Get(codId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*code), nil
}
