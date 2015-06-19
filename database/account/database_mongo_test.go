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

package account

import (
	"github.com/realglobe-Inc/edo-lib/base64url"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"testing"
	"time"
)

// テストするなら、mongodb を立てる必要あり。
// 立ってなかったらテストはスキップ。
var monPool *mgo.Session

func init() {
	if monPool == nil {
		monPool, _ = mgo.DialWithTimeout("localhost", time.Minute)
	}
}

const (
	test_coll = "test-collection"
)

func TestMongoDb(t *testing.T) {
	if monPool == nil {
		t.SkipNow()
	}

	test_db := "test-db-" + strconv.FormatInt(time.Now().UnixNano(), 16)
	conn := monPool.New()
	defer conn.Close()
	if err := conn.DB(test_db).C(test_coll).Insert(bson.M{
		"id":       test_id,
		"username": test_name,
		"authenticator": bson.M{
			"algorithm": "pbkdf2:sha256:1000",
			"salt":      base64url.EncodeToString(test_salt),
			"hash":      base64url.EncodeToString(test_pbkdf2Hash),
		},
		"pds": test_pds,
	}); err != nil {
		t.Fatal(err)
	}
	defer conn.DB(test_db).DropDatabase()

	testDb(t, NewMongoDb(monPool, test_db, test_coll))
}
