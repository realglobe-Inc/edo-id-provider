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

package key

import (
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/go-lib/erro"
)

// redis による自身の鍵のキャッシュ。
type redisCache struct {
	base  Db
	pool  *redis.Pool
	tag   string
	expIn time.Duration
}

func NewRedisCache(base Db, pool *redis.Pool, tag string, expIn time.Duration) Db {
	return &redisCache{
		base:  base,
		pool:  pool,
		tag:   tag,
		expIn: expIn,
	}
}

func (this *redisCache) Get() ([]jwk.Key, error) {
	conn := this.pool.Get()
	defer conn.Close()

	if data, err := redis.Bytes(conn.Do("GET", this.tag)); err != nil {
		if err != redis.ErrNil {
			log.Warn(erro.Wrap(err))
			// キャッシュが取れなくても諦めない。
		}
	} else {
		// キャッシュされてた。
		var ma []map[string]interface{}
		if err := json.Unmarshal(data, &ma); err != nil {
			return nil, erro.Wrap(err)
		}
		keys := []jwk.Key{}
		for _, m := range ma {
			key, err := jwk.FromMap(m)
			if err != nil {
				return nil, erro.Wrap(err)
			}
			keys = append(keys, key)
		}
		return keys, nil
	}

	// キャッシュされてなかった。
	keys, err := this.base.Get()
	if err != nil {
		return nil, erro.Wrap(err)
	} else if keys == nil {
		return nil, nil
	}

	// キャッシュする。
	ma := []map[string]interface{}{}
	for _, key := range keys {
		ma = append(ma, key.ToMap())
	}
	if data, err := json.Marshal(ma); err != nil {
		log.Warn(erro.Wrap(err))
		// キャッシュできなくても諦めない。
	} else if _, err := conn.Do("SET", this.tag, data, "PX", int64(this.expIn/time.Millisecond)); err != nil {
		log.Warn(erro.Wrap(err))
		// キャッシュできなくても諦めない。
	}

	return keys, nil
}
