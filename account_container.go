package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

type accountContainer interface {
	get(accId string) (*account, error)
	getByName(accName string) (*account, error)

	close() error
}

type accountContainerImpl struct {
	idToAcc   driver.KeyValueStore
	nameToAcc driver.KeyValueStore
}

func (this *accountContainerImpl) get(accId string) (*account, error) {
	val, _, err := this.idToAcc.Get(accId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*account), nil
}

func (this *accountContainerImpl) getByName(accName string) (*account, error) {
	val, _, err := this.nameToAcc.Get(accName, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*account), nil
}

func (this *accountContainerImpl) close() error {
	if err := this.idToAcc.Close(); err != nil {
		return erro.Wrap(err)
	} else if err := this.nameToAcc.Close(); err != nil {
		return erro.Wrap(err)
	}
	return nil
}
