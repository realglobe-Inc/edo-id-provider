package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib/erro"
)

type taContainer interface {
	get(taId string) (*ta, error)

	close() error
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

func (this *taContainerImpl) close() error {
	return this.base.Close()
}
