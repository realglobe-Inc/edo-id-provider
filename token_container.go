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

type tokenContainer interface {
	newId() (id string, err error)
	get(tokId string) (*token, error)
	put(tok *token) error

	close() error
}

type tokenContainerImpl struct {
	base driver.VolatileKeyValueStore

	idGenerator
	// 有効期限が切れてからも保持する期間。
	savDur time.Duration
}

func (this *tokenContainerImpl) get(tokId string) (*token, error) {
	val, _, err := this.base.Get(tokId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*token), nil
}

func (this *tokenContainerImpl) put(tok *token) error {
	if _, err := this.base.Put(tok.id(), tok, tok.expirationDate().Add(this.savDur)); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *tokenContainerImpl) close() error {
	return this.base.Close()
}
