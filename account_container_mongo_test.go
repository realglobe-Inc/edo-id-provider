package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
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

func TestMongoAccountContainer(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	sess, err := mgo.Dial(mongoAddr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.DB(testLabel).C("edo-id-provider").Upsert(bson.M{"id": testAcc.Id},
		&accountIntermediate{testAcc.Id, testAcc.Name, testAcc.Passwd, time.Now(), "xyz"}); err != nil {
		t.Fatal(err)
	}
	defer sess.DB(testLabel).C("edo-id-provider").DropCollection()

	testAccountContainer(t, newMongoAccountContainer(mongoAddr, testLabel, "edo-id-provider", 0, 0))
}
