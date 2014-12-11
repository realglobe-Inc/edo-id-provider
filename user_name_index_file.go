package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func NewFileUserNameIndex(path string, expiDur time.Duration) UserNameIndex {
	return newUserNameIndex(driver.NewFileKeyValueStore(path, keyToJsonPath, jsonPathToKey, json.Marshal, jsonUnmarshal, expiDur, expiDur))
}
