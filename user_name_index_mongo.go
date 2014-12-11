package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func NewMongoUserNameIndex(url, dbName, collName string, staleDur, expiDur time.Duration) UserNameIndex {
	return newUserNameIndex(driver.NewMongoKeyValueStore(url, dbName, collName, nil, nil, nil, staleDur, expiDur))
}
