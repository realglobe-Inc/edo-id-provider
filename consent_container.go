package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

type consentContainer interface {
	// 同意の得られている scope とクレームを返す。
	get(accId, taId string) (scops, clms map[string]bool, err error)
	// 同意を設定する。
	put(accId, taId string, consScops, consClms, denyScops, denyClms map[string]bool) error
}

type consentContainerImpl struct {
	base driver.KeyValueStore

	getKey func(accId, taId string) (key string)
}

func (this *consentContainerImpl) get(accId, taId string) (scope, clms map[string]bool, err error) {
	val, _, err := this.base.Get(this.getKey(accId, taId), nil)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil, nil
	} else if cons, ok := val.(*consent); !ok {
		return nil, nil, nil
	} else {
		return cons.Scops, cons.Clms, nil
	}
}

func (this *consentContainerImpl) put(accId, taId string, consScops, consClms, denyScops, denyClms map[string]bool) error {
	key := this.getKey(accId, taId)

	var scops, clms map[string]bool
	if val, _, err := this.base.Get(this.getKey(accId, taId), nil); err != nil {
		return erro.Wrap(err)
	} else if val == nil {
		scops, clms = map[string]bool{}, map[string]bool{}
	} else if cons, ok := val.(*consent); !ok {
		scops, clms = map[string]bool{}, map[string]bool{}
	} else {
		scops, clms = cons.Scops, cons.Clms
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

	if _, err := this.base.Put(key, &consent{scops, clms}); err != nil {
		return erro.Wrap(err)
	}
	return nil
}