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

type memorySessionContainer sessionContainerImpl

// スレッドセーフ。
func newMemorySessionContainer(minIdLen int, procId string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return &memorySessionContainer{
		driver.NewMemoryConcurrentVolatileKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
	}
}

func (this *memorySessionContainer) put(sess *session) error {
	return ((*sessionContainerImpl)(this)).put(sess.copy())
}

func (this *memorySessionContainer) get(sessId string) (*session, error) {
	sess, err := ((*sessionContainerImpl)(this)).get(sessId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if sess == nil {
		return nil, nil
	}
	return sess.copy(), nil
}

func (this *memorySessionContainer) close() error {
	return ((*sessionContainerImpl)(this)).close()
}
