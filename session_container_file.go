package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

func unmarshalSession(data []byte) (val interface{}, err error) {
	var res session
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

// スレッドセーフ。
func newFileSessionContainer(idLen int, expiDur time.Duration, path, expiPath string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return &sessionContainerImpl{
		idLen, expiDur,
		driver.NewFileTimeLimitedKeyValueStore(path, expiPath,
			keyToJsonPath, nil, json.Marshal, unmarshalSession,
			caStaleDur, caExpiDur),
	}
}
