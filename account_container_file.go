package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

func unmarshalAccount(data []byte) (interface{}, error) {
	var acc account
	if err := json.Unmarshal(data, &acc); err != nil {
		return nil, erro.Wrap(err)
	}
	return &acc, nil
}

// スレッドセーフ。
func newFileAccountContainer(idPath, namePath string, staleDur, expiDur time.Duration) accountContainer {
	return &accountContainerImpl{
		driver.NewFileListedKeyValueStore(idPath, keyToEscapedJsonPath, nil, nil, unmarshalAccount, staleDur, expiDur),
		driver.NewFileListedKeyValueStore(namePath, keyToEscapedJsonPath, nil, nil, unmarshalAccount, staleDur, expiDur),
	}
}
