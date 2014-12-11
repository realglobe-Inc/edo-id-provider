package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func NewMongoUserAttributeRegistry(url, dbName, collName string, staleDur, expiDur time.Duration) UserAttributeRegistry {
	return newUserAttributeRegistry(driver.NewMongoKeyValueStore(url, dbName, collName, nil, nil, nil, staleDur, expiDur))
}
