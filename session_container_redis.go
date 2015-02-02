package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"strconv"
	"time"
)

func getSessionStamp(val interface{}) *driver.Stamp {
	sess, _ := val.(*session)
	upd := sess.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

func newRedisSessionContainer(minIdLen int, procId string,
	pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return &sessionContainerImpl{
		driver.NewRedisTimeLimitedKeyValueStore(pool, tag+":", json.Marshal, unmarshalSession, getSessionStamp, caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
	}
}
