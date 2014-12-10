package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

// テストするなら、ローカルにデフォルトポートで mongodb をたてる必要あり。
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

func TestMongoTaExplorer(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg, err := NewMongoTaExplorer(mongoAddr, testLabel, "ta-explorer", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.(*taExplorer).base.(driver.MongoKeyValueStore).Clear()

	if _, err := reg.(*taExplorer).base.Put("list", bson.M{testUri: testServUuid}); err != nil {
		t.Fatal(err)
	}

	testTaExplorer(t, reg)
}

func TestMongoTaExplorerStamp(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg, err := NewMongoTaExplorer(mongoAddr, testLabel, "ta-explorer", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.(*taExplorer).base.(driver.MongoKeyValueStore).Clear()

	if _, err := reg.(*taExplorer).base.Put("list", bson.M{testUri: testServUuid}); err != nil {
		t.Fatal(err)
	}

	testTaExplorerStamp(t, reg)
}
