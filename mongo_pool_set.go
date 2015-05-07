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
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2"
	"time"
)

// mongodb のコネクションプール集。
type mongoPoolSet struct {
	timeout time.Duration
	pools   map[string]*mgo.Session
}

// timeout: 接続を待つ期間。
func newMongoPoolSet(timeout time.Duration) *mongoPoolSet {
	return &mongoPoolSet{
		timeout: timeout,
		pools:   map[string]*mgo.Session{},
	}
}

func (this *mongoPoolSet) get(addr string) (*mgo.Session, error) {
	pool := this.pools[addr]
	if pool != nil {
		return pool, nil
	}

	pool, err := mgo.DialWithTimeout(addr, this.timeout)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	pool.SetSyncTimeout(this.timeout)
	this.pools[addr] = pool
	return pool, nil
}

func (this *mongoPoolSet) close() {
	for _, pool := range this.pools {
		pool.Close()
	}
}
