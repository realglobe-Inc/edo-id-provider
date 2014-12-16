package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

type account struct {
	// IdP 内で一意かつ変更されることのない ID。
	Id string `json:"id"`
	// IdP 内で一意のログイン ID。
	Name string `json:"name"`
	// パスワード。
	Passwd string `json:"passwd"`
}

type accountContainer interface {
	get(accId string) (*account, error)
	getByName(accName string) (*account, error)
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
