package main

import (
	"github.com/realglobe-Inc/edo-lib/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

func readConsent(query *mgo.Query) (interface{}, error) {
	var res consent
	if err := query.One(&res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func getConsentStamp(val interface{}) *driver.Stamp {
	cons, _ := val.(*consent)
	upd := cons.updateDate()
	return &driver.Stamp{Date: upd, Digest: strconv.FormatInt(upd.UnixNano(), 16)}
}

type mongoConsentContainer struct {
	base driver.MongoNKeyValueStore
}

// スレッドセーフ。
func newMongoConsentContainer(url, dbName, collName string, staleDur, expiDur time.Duration) consentContainer {
	return &mongoConsentContainer{
		driver.NewMongoNKeyValueStore(url, dbName, collName,
			[]string{"account", "client_id"}, nil, nil, readConsent, getConsentStamp,
			staleDur, expiDur),
	}
}

func (this *mongoConsentContainer) get(accId, taId string) (scope, clms map[string]bool, err error) {
	val, _, err := this.base.NGet(bson.M{"account": accId, "client_id": taId}, nil)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil, nil
	} else if cons, ok := val.(*consent); !ok {
		return nil, nil, nil
	} else {
		return cons.scopes(), cons.claims(), nil
	}
}

func (this *mongoConsentContainer) put(accId, taId string, consScops, consClms, denyScops, denyClms map[string]bool) error {
	var scops, clms map[string]bool

	val, _, err := this.base.NGet(bson.M{"account": accId, "client_id": taId}, nil)
	if err != nil {
		return erro.Wrap(err)
	} else if val == nil {
		scops, clms = map[string]bool{}, map[string]bool{}
	} else if cons, ok := val.(*consent); !ok {
		scops, clms = map[string]bool{}, map[string]bool{}
	} else {
		scops, clms = cons.scopes(), cons.claims()
	}

	for scop := range consScops {
		scops[scop] = true
	}
	for clm := range consClms {
		clms[clm] = true
	}
	for scop := range denyScops {
		delete(scops, scop)
	}
	for clm := range denyClms {
		delete(clms, clm)
	}

	if _, err := this.base.NPut(bson.M{"account": accId, "client_id": taId}, newConsent(accId, taId, scops, clms)); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *mongoConsentContainer) close() error {
	return this.base.Close()
}
