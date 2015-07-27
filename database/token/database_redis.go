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

package token

import (
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/go-lib/erro"
)

const (
	tagDate = "date"
)

type redisDb struct {
	pool *redis.Pool
	tag  string
}

// redis によるアクセストークン情報の格納庫。
// 本体と別に更新日時を、日時タグと ID をキーにして、Unix 時間 (ナノ秒) で格納。
// RFC3339 じゃない理由は同値比較でタイムゾーンを考えたくないから。
func NewRedisDb(pool *redis.Pool, tag string) Db {
	return &redisDb{
		pool: pool,
		tag:  tag,
	}
}

// 取得。
func (this *redisDb) Get(id string) (*Element, error) {
	conn := this.pool.Get()
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", this.tag+id))
	if err != nil {
		if err == redis.ErrNil {
			// 無かった。
			return nil, nil
		}
		return nil, erro.Wrap(err)
	}

	var elem Element
	if err := json.Unmarshal(data, &elem); err != nil {
		return nil, erro.Wrap(err)
	}

	return &elem, nil
}

// 保存。
func (this *redisDb) Save(elem *Element, exp time.Time) error {
	conn := this.pool.Get()
	defer conn.Close()

	data, err := json.Marshal(elem)
	if err != nil {
		return erro.Wrap(err)
	}
	expIn := int64(exp.Sub(time.Now()) / time.Millisecond)

	conn.Send("MULTI")
	conn.Send("SET", this.tag+elem.Id(), data, "PX", expIn)
	conn.Send("SET", this.tag+tagDate+elem.Id(), elem.Date().UnixNano(), "PX", expIn)
	if _, err := conn.Do("EXEC"); err != nil {
		return erro.Wrap(err)
	}

	return nil
}

var replaceScript = redis.NewScript(2, `
if redis.call("get", KEYS[2]) ~= ARGV[1] then
   return "invalid date"
end
local exp_in = redis.call("pttl", KEYS[1])
redis.call("set", KEYS[1], ARGV[2], "PX", exp_in)
redis.call("set", KEYS[2], ARGV[3], "PX", exp_in)
return ""
`)

// 上書き。
func (this *redisDb) Replace(elem *Element, savedDate time.Time) (ok bool, err error) {
	conn := this.pool.Get()
	defer conn.Close()

	data, err := json.Marshal(elem)
	if err != nil {
		return false, erro.Wrap(err)
	}

	res, err := redis.String(replaceScript.Do(conn, this.tag+elem.Id(), this.tag+tagDate+elem.Id(), savedDate.UnixNano(), data, elem.Date().UnixNano()))
	if err != nil {
		return false, erro.Wrap(err)
	} else if res != "" {
		if res == "invalid date" {
			return false, nil
		}
		return false, erro.New(res)
	}

	return true, nil
}
