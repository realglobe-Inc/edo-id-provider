package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

func unmarshalToken(data []byte) (val interface{}, err error) {
	var res token
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

// スレッドセーフ。
func newFileTokenContainer(idLen int, path, expiPath string, caStaleDur, caExpiDur time.Duration) tokenContainer {
	return &tokenContainerImpl{
		idLen,
		driver.NewFileTimeLimitedKeyValueStore(path, expiPath,
			keyToJsonPath, nil, json.Marshal, unmarshalToken,
			caStaleDur, caExpiDur),
	}
}
