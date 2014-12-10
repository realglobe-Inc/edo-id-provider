package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

// {
//   "service": {
//     "public_key": "XXXXX"
//   }
// }
func webServicePublicKeyUnmarshal(data []byte) (interface{}, error) {
	var res struct {
		Service struct {
			Public_key string
		}
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return util.ParseRsaPublicKey(res.Service.Public_key)
}

// スレッドセーフ。
func NewWebTaKeyProvider(prefix string) TaKeyProvider {
	return newTaKeyProvider(driver.NewWebKeyValueStore(prefix, nil, webServicePublicKeyUnmarshal))
}
