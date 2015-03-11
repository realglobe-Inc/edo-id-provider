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

func readAccount(query *mgo.Query) (interface{}, error) {
	var res account
	if err := query.One(&res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func getAccountStamp(val interface{}) *driver.Stamp {
	acc := val.(*account)
	upd := acc.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

type mongoAccountContainer struct {
	idToAcc   driver.KeyValueStore
	nameToAcc driver.KeyValueStore
}

// スレッドセーフ。
func newMongoAccountContainer(url, dbName, collName string, staleDur, expiDur time.Duration) accountContainer {
	return &mongoAccountContainer{
		driver.NewMongoKeyValueStore(url, dbName, collName,
			"id", nil, nil, readAccount, getAccountStamp,
			staleDur, expiDur),
		driver.NewMongoKeyValueStore(url, dbName, collName,
			"username", nil, nil, readAccount, getAccountStamp,
			staleDur, expiDur),
	}
}

func (this *mongoAccountContainer) get(accId string) (*account, error) {
	val, _, err := this.idToAcc.Get(accId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*account), nil
}

func (this *mongoAccountContainer) getByName(accName string) (*account, error) {
	val, _, err := this.nameToAcc.Get(accName, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*account), nil
}

func (this *mongoAccountContainer) close() error {
	if err := this.idToAcc.Close(); err != nil {
		return erro.Wrap(err)
	} else if err := this.nameToAcc.Close(); err != nil {
		return erro.Wrap(err)
	}
	return nil
}
