package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestMongoTaContainer(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	sess, err := mgo.Dial(mongoAddr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.DB(testLabel).C("edo-id-provider").Upsert(bson.M{"id": testTa.id()},
		taToIntermediate(testTa)); err != nil {
		t.Fatal(err)
	}
	defer sess.DB(testLabel).C("edo-id-provider").DropCollection()

	testTaContainer(t, newMongoTaContainer(mongoAddr, testLabel, "edo-id-provider", testStaleDur, testCaExpiDur))
}
