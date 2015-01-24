package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"strconv"
	"time"
)

func getDummyStamp(val interface{}) *driver.Stamp {
	return &driver.Stamp{}
}

func getCodeStamp(val interface{}) *driver.Stamp {
	cod, _ := val.(*code)
	upd := cod.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

func newRedisCodeContainer(minIdLen int, expiDur time.Duration, pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &codeContainerImpl{
		driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalCode, getCodeStamp, caStaleDur, caExpiDur),
		newIdGenerator(minIdLen),
		expiDur,
	}
}
