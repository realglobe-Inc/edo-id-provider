package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
)

func TestMongoUserNameIndex(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg := NewMongoUserNameIndex(mongoAddr, testLabel, "user_ids", 0)
	defer reg.(*userNameIndex).base.(driver.MongoKeyValueStore).Clear()

	if _, err := reg.(*userNameIndex).base.Put(testUsrName, testUsrUuid); err != nil {
		t.Fatal(err)
	}

	testUserNameIndex(t, reg)
}
