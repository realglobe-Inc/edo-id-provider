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
		Value map[string]string
		Stamp *driver.Stamp
	}
	if err := query.One(&res); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	return res.Value, res.Stamp, nil
}

// スレッドセーフ。
func NewMongoTaExplorer(url, dbName, collName string, expiDur time.Duration) (TaExplorer, error) {
	base, err := driver.NewMongoKeyValueStore(url, dbName, collName, expiDur)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	base.SetUnmarshal(containerToTaExplorerTree)
	base.SetTake(containerMongoTake)
	return newTaExplorer(base), nil
}
