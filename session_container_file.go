package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

func marshalSession(val interface{}) (data []byte, err error) {
	return json.Marshal(sessionToIntermediate(val.(*session)))
}

func unmarshalSession(data []byte) (val interface{}, err error) {
	var res sessionIntermediate
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return intermediateToSession(&res), nil
}

// スレッドセーフ。
func newFileSessionContainer(idLen int, expiDur time.Duration, path, expiPath string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return &sessionContainerWrapper{idLen, expiDur,
		&sessionContainerImpl{
			driver.NewFileTimeLimitedKeyValueStore(path, expiPath,
				keyToJsonPath, nil, marshalSession, unmarshalSession,
				caStaleDur, caExpiDur),
		}}
}
