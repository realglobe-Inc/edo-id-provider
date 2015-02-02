package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

func unmarshalCode(data []byte) (val interface{}, err error) {
	var res code
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

// スレッドセーフ。
func newFileCodeContainer(minIdLen int, procId string, savDur time.Duration, path, expiPath string, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &codeContainerImpl{
		driver.NewFileVolatileKeyValueStore(path, expiPath,
			keyToJsonPath, nil, json.Marshal, unmarshalCode,
			caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
	}
}
