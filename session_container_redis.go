package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

func newRedisSessionContainer(idLen int, expiDur time.Duration, pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return &sessionContainerWrapper{idLen, expiDur,
		&sessionContainerImpl{
			driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalSession, getDummyStamp, caStaleDur, caExpiDur),
		}}
}
