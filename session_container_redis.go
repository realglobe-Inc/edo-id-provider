package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

func newRedisSessionContainer(minIdLen int, pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return newSessionContainerImpl(
		driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalSession, getDummyStamp, caStaleDur, caExpiDur),
		minIdLen)
}
