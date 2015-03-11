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
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo-lib/driver"
	"time"
)

// テストするなら、redis をたてる必要あり。
var redisAddr = ":6379"
var testRedisPool *redis.Pool

func init() {
	if redisAddr != "" {
		// 実際にサーバーが立っているかどうか調べる。
		// 立ってなかったらテストはスキップ。
		conn, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			redisAddr = ""
		} else {
			conn.Close()
		}
	}

	if redisAddr != "" {
		testRedisPool = driver.NewRedisPool(redisAddr, 2, time.Second)
	}
}
