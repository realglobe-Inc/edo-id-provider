package main

import (
	"github.com/realglobe-Inc/edo/driver"
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
func newMongoTaContainer(url, dbName, collName string, staleDur, expiDur time.Duration) taContainer {
	return &taContainerImpl{driver.NewMongoKeyValueStore(url, dbName, collName,
		"id", nil, nil, readTa, getTaStamp,
		staleDur, expiDur)}
}
