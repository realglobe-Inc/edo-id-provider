package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

func getDummyStamp(val interface{}) *driver.Stamp {
	return &driver.Stamp{}
}

func newRedisCodeContainer(idLen int, expiDur time.Duration, selfId string, pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &codeContainerImpl{
		idLen, expiDur, selfId,
		driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalCode, getDummyStamp, caStaleDur, caExpiDur),
	}
}
