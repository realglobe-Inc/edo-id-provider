package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"strconv"
	"time"
)

func getTokenStamp(val interface{}) *driver.Stamp {
	tok, _ := val.(*token)
	upd := tok.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

func newRedisTokenContainer(minIdLen int, savDur time.Duration, pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) tokenContainer {
	return &tokenContainerImpl{
		driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalToken, getTokenStamp, caStaleDur, caExpiDur),
		newIdGenerator(minIdLen),
		savDur,
	}
}
