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
