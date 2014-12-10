package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
)

func TestMongoUserAttributeRegistry(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg, err := NewMongoUserAttributeRegistry(mongoAddr, testLabel, "user_attributes", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer reg.(*userAttributeRegistry).base.(driver.MongoKeyValueStore).Clear()

	if _, err := reg.(*userAttributeRegistry).base.Put(testUsrUuid+"/"+testAttrName, testAttr); err != nil {
		t.Fatal(err)
	}

	testUserAttributeRegistry(t, reg)
}
