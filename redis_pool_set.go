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
	"time"
)

// redis のコネクションプール集。
type redisPoolSet struct {
	timeout time.Duration
	n       int
	keep    time.Duration
	pools   map[string]*redis.Pool
}

// timeout: 通信タイムアウト。
// n: プール毎に確保する接続数。
// keep: 確保する期間。
func newRedisPoolSet(timeout time.Duration, n int, keep time.Duration) *redisPoolSet {
	return &redisPoolSet{
		n:     n,
		keep:  keep,
		pools: map[string]*redis.Pool{},
	}
}

func (this *redisPoolSet) get(addr string) *redis.Pool {
	pool := this.pools[addr]
	if pool != nil {
		return pool
	}

	pool = &redis.Pool{
		MaxIdle:     this.n,
		IdleTimeout: this.keep,
		Dial: func() (redis.Conn, error) {
			return redis.DialTimeout("tcp", addr, this.timeout, this.timeout, this.timeout)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	this.pools[addr] = pool
	return pool
}

func (this *redisPoolSet) close() {
	for _, pool := range this.pools {
		pool.Close()
	}
}
