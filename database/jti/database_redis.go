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

package jti

import (
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

const (
	tagSep = ":"
)

// redis による JWT の ID の格納庫。
// JWT の iss と jti をキーとして、有効期限をデータの有効期限として格納する。
// データんの中身は無い。
type redisDb struct {
	pool *redis.Pool
	tag  string
}

func NewRedisDb(pool *redis.Pool, tag string) Db {
	return &redisDb{
		pool: pool,
		tag:  tag,
	}
}

func (this *redisDb) SaveIfAbsent(elem *Element) (ok bool, err error) {
	conn := this.pool.Get()
	defer conn.Close()

	expIn := int64(elem.Expires().Sub(time.Now()) / time.Millisecond)

	res, err := conn.Do("Set", this.tag+elem.Issuer()+tagSep+elem.Id(), "", "PX", expIn, "NX")
	if err != nil {
		return false, erro.Wrap(err)
	} else if res == nil {
		// 既にあった。
		return false, nil
	}

	return true, nil
}
