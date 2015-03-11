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

type memoryCodeContainer codeContainerImpl

// スレッドセーフ。
func newMemoryCodeContainer(minIdLen int, procId string, savDur, ticExpDur, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &memoryCodeContainer{
		driver.NewMemoryConcurrentVolatileKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
		ticExpDur,
	}
}

func (this *memoryCodeContainer) get(codId string) (*code, error) {
	cod, err := ((*codeContainerImpl)(this)).get(codId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if cod == nil {
		return nil, nil
	}
	return cod.copy(), nil
}

func (this *memoryCodeContainer) put(cod *code) error {
	return ((*codeContainerImpl)(this)).put(cod.copy())
}

func (this *memoryCodeContainer) getAndSetEntry(codId string) (cod *code, tic string, err error) {
	cod, tic, err = ((*codeContainerImpl)(this)).getAndSetEntry(codId)
	if err != nil {
		return nil, "", erro.Wrap(err)
	} else if cod == nil {
		return nil, tic, nil
	}
	return cod.copy(), tic, nil
}

func (this *memoryCodeContainer) putIfEntered(cod *code, tic string) (ok bool, err error) {
	return ((*codeContainerImpl)(this)).putIfEntered(cod.copy(), tic)
}

func (this *memoryCodeContainer) close() error {
	return ((*codeContainerImpl)(this)).close()
}
