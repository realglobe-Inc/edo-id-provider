// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	if _, err := sess.DB(testLabel).C("edo-id-provider").Upsert(bson.M{"id": testTa.id()}, testTa); err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	defer sess.DB(testLabel).C("edo-id-provider").DropCollection()

	testTaContainer(t, newMongoTaContainer(mongoAddr, testLabel, "edo-id-provider", testStaleDur, testCaExpiDur))
}
