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
	"time"
)

type codeContainer interface {
	newId() (id string, err error)
	get(codId string) (*code, error)
	put(cod *code) error
	// tic は書き込み券。
	getAndSetEntry(codId string) (cod *code, tic string, err error)
	// tic が最後に発行された書き込み券だった場合のみ書き込まれ、ok が true になる。
	putIfEntered(cod *code, tic string) (ok bool, err error)

	close() error
}

type codeContainerImpl struct {
	base driver.ConcurrentVolatileKeyValueStore

	idGenerator
	// 有効期限が切れてからも保持する期間。
	savDur time.Duration
	// 書き込み券の有効期間。
	ticExpDur time.Duration
}

func (this *codeContainerImpl) get(codId string) (*code, error) {
	val, _, err := this.base.Get(codId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*code), nil
}

func (this *codeContainerImpl) put(cod *code) error {
	if _, err := this.base.Put(cod.id(), cod, cod.expirationDate().Add(this.savDur)); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *codeContainerImpl) getAndSetEntry(codId string) (cod *code, tic string, err error) {
	tic, _ = this.idGenerator.id(0)
	val, _, err := this.base.GetAndSetEntry(codId, nil, codId+":entry", tic, time.Now().Add(this.ticExpDur))
	if err != nil {
		return nil, "", erro.Wrap(err)
	} else if val == nil {
		return nil, tic, nil
	}
	return val.(*code), tic, nil
}

func (this *codeContainerImpl) putIfEntered(cod *code, tic string) (ok bool, err error) {
	ok, _, err = this.base.PutIfEntered(cod.id(), cod, cod.expirationDate().Add(this.savDur), cod.id()+":entry", tic)
	if err != nil {
		return false, erro.Wrap(err)
	}
	return ok, nil
}

func (this *codeContainerImpl) close() error {
	return this.base.Close()
}
