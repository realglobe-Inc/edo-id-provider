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
)

type consentContainer interface {
	// 同意の得られている scope とクレームを返す。
	get(accId, taId string) (scops, clms map[string]bool, err error)
	// 同意を設定する。
	put(accId, taId string, consScops, consClms, denyScops, denyClms map[string]bool) error

	close() error
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
		return cons.scopes(), cons.claims(), nil
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

	if _, err := this.base.Put(key, newConsent(accId, taId, scops, clms)); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *consentContainerImpl) close() error {
	return this.base.Close()
}
