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
	"github.com/realglobe-Inc/edo-lib/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2"
	"strconv"
	"time"
)

func readTa(query *mgo.Query) (interface{}, error) {
	var res ta
	if err := query.One(&res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func getTaStamp(val interface{}) *driver.Stamp {
	t := val.(*ta)
	upd := t.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

// スレッドセーフ。
func newMongoTaContainer(sess *mgo.Session, dbName, collName string, staleDur, expiDur time.Duration) taContainer {
	return &taContainerImpl{driver.NewMongoKeyValueStore(sess, dbName, collName,
		"id", nil, nil, readTa, getTaStamp,
		staleDur, expiDur)}
}
