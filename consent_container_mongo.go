package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

type consentIntermediate struct {
	Acc string `bson:"account"`
	Ta  string `bson:"ta"`

	Cons consent `bson:"consent,omitempty"`

	Digest string    `bson:"digest"`
	Date   time.Time `bson:"date"`
}

func readConsent(query *mgo.Query) (interface{}, error) {
	var res consentIntermediate
	if err := query.One(&res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func getConsentStamp(val interface{}) *driver.Stamp {
	cons, _ := val.(*consentIntermediate)
	return &driver.Stamp{Date: cons.Date, Digest: cons.Digest}
}

type mongoConsentContainer struct {
	base driver.MongoNKeyValueStore

	ser int64
}

// スレッドセーフ。
func newMongoConsentContainer(url, dbName, collName string, staleDur, expiDur time.Duration) consentContainer {
	return &mongoConsentContainer{
		driver.NewMongoNKeyValueStore(url, dbName, collName,
			[]string{"account", "ta"}, nil, nil, readConsent, getConsentStamp,
			staleDur, expiDur),
		rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
	}
}

func (this *mongoConsentContainer) get(accId, taId string) (scope, clms map[string]bool, err error) {
	val, _, err := this.base.NGet(bson.M{"account": accId, "ta": taId}, nil)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil, nil
	} else if cons, ok := val.(*consentIntermediate); !ok {
		return nil, nil, nil
	} else {
		return cons.Cons.Scops, cons.Cons.Clms, nil
	}
}

func (this *mongoConsentContainer) put(accId, taId string, consScops, consClms, denyScops, denyClms map[string]bool) error {
	var scops, clms map[string]bool

	val, _, err := this.base.NGet(bson.M{"account": accId, "ta": taId}, nil)
	if err != nil {
		return erro.Wrap(err)
	} else if val == nil {
		scops, clms = map[string]bool{}, map[string]bool{}
	} else if cons, ok := val.(*consentIntermediate); !ok {
		scops, clms = map[string]bool{}, map[string]bool{}
	} else {
		scops, clms = cons.Cons.Scops, cons.Cons.Clms
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

	cons := &consentIntermediate{
		Acc:    accId,
		Ta:     taId,
		Cons:   consent{scops, clms},
		Digest: strconv.FormatInt(atomic.AddInt64(&this.ser, 1), 10),
		Date:   time.Now(),
	}
	if _, err := this.base.NPut(bson.M{"account": accId, "ta": taId}, cons); err != nil {
		return erro.Wrap(err)
	}
	return nil
}
