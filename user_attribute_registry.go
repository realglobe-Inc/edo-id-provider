package main

import (
	"github.com/realglobe-Inc/edo/driver"
)

type UserAttributeRegistry interface {
	// ユーザーの属性を返す。
	UserAttribute(usrUuid, attrName string, caStmp *driver.Stamp) (usrAttr interface{}, newCaStmp *driver.Stamp, err error)
}

// 骨組み。
type userAttributeRegistry struct {
	base driver.KeyValueStore
}

func newUserAttributeRegistry(base driver.KeyValueStore) *userAttributeRegistry {
	return &userAttributeRegistry{base}
}

func (reg *userAttributeRegistry) UserAttribute(usrUuid, attrName string, caStmp *driver.Stamp) (usrAttr interface{}, newCaStmp *driver.Stamp, err error) {
	return reg.base.Get(usrUuid+"/"+attrName, caStmp)
}
