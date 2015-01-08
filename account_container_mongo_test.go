package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

func TestMongoAccountContainer(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	sess, err := mgo.Dial(mongoAddr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.DB(testLabel).C("edo-id-provider").Upsert(bson.M{"id": testAcc.id()}, newAccount(map[string]interface{}{
		"id":       testAcc.id(),
		"username": testAcc.name(),
		"password": testAcc.password(),
		"date":     time.Now(),
		"digest":   "xyz"})); err != nil {
		t.Fatal(err)
	}
	defer sess.DB(testLabel).C("edo-id-provider").DropCollection()

	testAccountContainer(t, newMongoAccountContainer(mongoAddr, testLabel, "edo-id-provider", 0, 0))
}
