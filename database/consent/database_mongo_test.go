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

package consent

import (
	"gopkg.in/mgo.v2"
	"strconv"
	"testing"
	"time"
)

// テストするなら、mongodb を立てる必要あり。
// 立ってなかったらテストはスキップ。
var monPool, _ = mgo.DialWithTimeout("localhost", time.Minute)

func init() {
	if monPool != nil {
		monPool.SetSyncTimeout(time.Minute)
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
	defer conn.DB(test_db).DropDatabase()

	testDb(t, NewMongoDb(monPool, test_db, test_coll))
}
