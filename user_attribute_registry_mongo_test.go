package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
)

func TestMongoUserAttributeRegistry(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg := NewMongoUserAttributeRegistry(mongoAddr, testLabel, "user_attributes", 0)
	defer reg.(*userAttributeRegistry).base.(driver.MongoKeyValueStore).Clear()

	if _, err := reg.(*userAttributeRegistry).base.Put(testUsrUuid+"/"+testAttrName, testAttr); err != nil {
		t.Fatal(err)
	}

	testUserAttributeRegistry(t, reg)
}
