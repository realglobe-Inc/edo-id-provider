package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func NewMongoUserAttributeRegistry(url, dbName, collName string, expiDur time.Duration) (UserAttributeRegistry, error) {
	return newUserAttributeRegistry(driver.NewMongoKeyValueStore(url, dbName, collName, nil, nil, nil, expiDur, expiDur)), nil
}
