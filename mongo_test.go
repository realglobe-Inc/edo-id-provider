package main

import (
	"gopkg.in/mgo.v2"
)

// テストするなら、mongodb をたてる必要あり。
var mongoAddr = "localhost"

func init() {
	if mongoAddr != "" {
		// 実際にサーバーが立っているかどうか調べる。
		// 立ってなかったらテストはスキップ。
		conn, err := mgo.Dial(mongoAddr)
		if err != nil {
			mongoAddr = ""
		} else {
			conn.Close()
		}
	}
}
