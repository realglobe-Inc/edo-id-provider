package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

// ユーザー名からユーザー UUID を引く。
type UserNameIndex interface {
	UserUuid(usrName string, caStmp *driver.Stamp) (usrUuid string, newCaStmp *driver.Stamp, err error)
}

// 骨組み。
type userNameIndex struct {
	base driver.KeyValueStore
}

func newUserNameIndex(base driver.KeyValueStore) *userNameIndex {
	return &userNameIndex{base}
}

func (reg *userNameIndex) UserUuid(usrName string, caStmp *driver.Stamp) (usrUuid string, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := reg.base.Get(usrName, caStmp)
	if err != nil {
		return "", nil, erro.Wrap(err)
	} else if value == nil || value == "" {
		return "", newCaStmp, nil
	}
	return value.(string), newCaStmp, nil
}
