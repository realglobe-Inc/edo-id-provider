package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"gopkg.in/mgo.v2"
	"time"
)

// map[string]string から taExplorerTree をつくる。
func containerToTaExplorerTree(data interface{}) (interface{}, error) {
	tree := newTaExplorerTree()
	tree.fromContainer(data.(map[string]string))
	return tree, nil
}

// mongodb から map[string]string を読み取る。
func containerMongoTake(query *mgo.Query) (interface{}, *driver.Stamp, error) {
	var res struct {
		V map[string]string
		S *driver.Stamp
	}
	if err := query.One(&res); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	return res.V, res.S, nil
}

// スレッドセーフ。
func NewMongoTaExplorer(url, dbName, collName string, expiDur time.Duration) (TaExplorer, error) {
	return newTaExplorer(driver.NewMongoKeyValueStore(url, dbName, collName, nil, containerToTaExplorerTree, containerMongoTake, expiDur, expiDur)), nil
}
