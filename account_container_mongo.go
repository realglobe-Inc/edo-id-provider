package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"gopkg.in/mgo.v2"
	"time"
)

type accountIntermediate struct {
	Id     string `bson:"id"`
	Name   string `bson:"name"`
	Passwd string `bson:"passwd"`

	Date   time.Time `bson:"date"`
	Digest string    `bson:"digest"`
}

func readAccountIntermediate(query *mgo.Query) (interface{}, error) {
	var res accountIntermediate
	if err := query.One(&res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func getAccountIntermediateStamp(val interface{}) *driver.Stamp {
	acc := val.(*accountIntermediate)
	return &driver.Stamp{Date: acc.Date, Digest: acc.Digest}
}

type mongoAccountContainer struct {
	idToAcc   driver.KeyValueStore
	nameToAcc driver.KeyValueStore
}

// スレッドセーフ。
func newMongoAccountContainer(url, dbName, collName string, staleDur, expiDur time.Duration) accountContainer {
	return &mongoAccountContainer{
		driver.NewMongoKeyValueStore(url, dbName, collName,
			"id", nil, nil, readAccountIntermediate, getAccountIntermediateStamp,
			staleDur, expiDur),
		driver.NewMongoKeyValueStore(url, dbName, collName,
			"name", nil, nil, readAccountIntermediate, getAccountIntermediateStamp,
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
	ai := val.(*accountIntermediate)
	return &account{ai.Id, ai.Name, ai.Passwd}, nil
}

func (this *mongoAccountContainer) getByName(nameId string) (*account, error) {
	val, _, err := this.nameToAcc.Get(nameId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	ai := val.(*accountIntermediate)
	return &account{ai.Id, ai.Name, ai.Passwd}, nil
}
