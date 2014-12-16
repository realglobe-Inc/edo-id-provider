package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"gopkg.in/mgo.v2"
	"time"
)

func readTaIntermediate(query *mgo.Query) (interface{}, error) {
	var res taIntermediate
	if err := query.One(&res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func getTaIntermediateStamp(val interface{}) *driver.Stamp {
	ti := val.(*taIntermediate)
	return &driver.Stamp{Date: ti.Date, Digest: ti.Digest}
}

type mongoTaContainer struct {
	base driver.KeyValueStore
}

// スレッドセーフ。
func newMongoTaContainer(url, dbName, collName string, staleDur, expiDur time.Duration) taContainer {
	return &mongoTaContainer{driver.NewMongoKeyValueStore(url, dbName, collName,
		"id", nil, nil, readTaIntermediate, getTaIntermediateStamp,
		staleDur, expiDur)}
}

func (this *mongoTaContainer) get(taId string) (*ta, error) {
	val, _, err := this.base.Get(taId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	ti := val.(*taIntermediate)
	return intermediateToTa(ti)
}
