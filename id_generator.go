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
	"github.com/realglobe-Inc/edo-lib/prand"
	"github.com/realglobe-Inc/edo-lib/secrand"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

// 安全な乱数が使えないときの代替。
var pr = prand.New(time.Minute)

// 長さを指定して ID 用のランダム文字列をつくる。
func newId(length int) string {
	id, err := secrand.String(length)
	if err != nil {
		log.Err(erro.Wrap(err))
		id = pr.String(length)
	}
	return id
}

// 長さを指定して ID 用のランダムバイト列をつくる。
func newIdBytes(length int) []byte {
	id, err := secrand.Bytes(length)
	if err != nil {
		log.Err(erro.Wrap(err))
		id = pr.Bytes(length)
	}
	return id
}
