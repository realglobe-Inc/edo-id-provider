package main

import (
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"gopkg.in/mgo.v2"
	"testing"
)

func init() {
	util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
}

func TestMongoConsentContainer(t *testing.T) {
	// ////////////////////////////////
	// util.SetupConsoleLog("github.com/realglobe-Inc", level.ALL)
	// defer util.SetupConsoleLog("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////
	if mongoAddr == "" {
		t.SkipNow()
	}

	sess, err := mgo.Dial(mongoAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.DB(testLabel).C("edo-id-provider").DropCollection()

	testConsentContainer(t, newMongoConsentContainer(mongoAddr, testLabel, "edo-id-provider", testStaleDur, testCaExpiDur))
}
