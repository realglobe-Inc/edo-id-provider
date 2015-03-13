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

type sessionContainer interface {
	newId() (id string, err error)
	put(sess *session) error
	get(sessId string) (*session, error)

	close() error
}

type sessionContainerImpl struct {
	base driver.VolatileKeyValueStore

	idGenerator
}

func (this *sessionContainerImpl) put(sess *session) error {
	if _, err := this.base.Put(sess.id(), sess, sess.expirationDate()); err != nil {
		return erro.Wrap(err)
	}

	return nil
}

func (this *sessionContainerImpl) get(sessId string) (*session, error) {
	val, _, err := this.base.Get(sessId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*session), nil
}

func (this *sessionContainerImpl) close() error {
	return this.base.Close()
}
