package main

import (
	"github.com/garyburd/redigo/redis"
)

// テストするなら、redis をたてる必要あり。
var redisAddr = ":6379"

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
}
