package main

import (
	"crypto"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

func newRedisTokenContainer(idLen int, selfId string, key crypto.PrivateKey, kid, alg string, idTokExpiDur time.Duration,
	pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) tokenContainer {
	return &tokenContainerImpl{
		idLen, selfId, key, kid, alg, idTokExpiDur,
		driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalToken, getDummyStamp, caStaleDur, caExpiDur),
	}
}
