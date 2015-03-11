package main

import (
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"gopkg.in/mgo.v2"
	"testing"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

func TestMongoConsentContainer(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////
	if mongoAddr == "" {
		t.SkipNow()
	}

	sess, err := mgo.Dial(mongoAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	defer sess.DB(testLabel).C("edo-id-provider").DropCollection()

	testConsentContainer(t, newMongoConsentContainer(mongoAddr, testLabel, "edo-id-provider", testStaleDur, testCaExpiDur))
}
