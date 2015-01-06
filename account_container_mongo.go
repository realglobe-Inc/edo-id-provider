package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"gopkg.in/mgo.v2"
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
	date, _ := acc.attribute("date").(time.Time)
	dig, _ := acc.attribute("digest").(string)
	return &driver.Stamp{Date: date, Digest: dig}
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
			"name", nil, nil, readAccount, getAccountStamp,
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
