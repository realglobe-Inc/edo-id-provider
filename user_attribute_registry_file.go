package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

// data を JSON として、encoding/json の標準データ型にデコードする。
func jsonUnmarshal(data []byte) (interface{}, error) {
	var res interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return res, nil
}

// スレッドセーフ。
func NewFileUserAttributeRegistry(path string, expiDur time.Duration) UserAttributeRegistry {
	return newUserAttributeRegistry(driver.NewFileKeyValueStore(path, keyToJsonPath, jsonPathToKey, json.Marshal, jsonUnmarshal, expiDur, expiDur))
}
