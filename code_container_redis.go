// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo-lib/driver"
	"strconv"
	"time"
)

func getCodeStamp(val interface{}) *driver.Stamp {
	cod, _ := val.(*code)
	upd := cod.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

func newRedisCodeContainer(minIdLen int, procId string, savDur, ticExpDur time.Duration, pool *redis.Pool, tag string, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &codeContainerImpl{
		driver.NewRedisConcurrentVolatileKeyValueStore(pool, tag+":", json.Marshal, unmarshalCode, getCodeStamp, caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
		ticExpDur,
	}
}
