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
	"github.com/realglobe-Inc/edo-lib/test"
	"testing"
	"time"
)

func TestRedisPoolSet(t *testing.T) {
	red, err := test.NewRedisServer()
	if err != nil {
		t.Fatal(err)
	} else if red == nil {
		t.SkipNow()
	}
	defer red.Close()
	addr := red.Address()

	red2, err := test.NewRedisServer()
	if err != nil {
		t.Fatal(err)
	} else if red == nil {
		t.SkipNow()
	}
	defer red2.Close()
	addr2 := red2.Address()

	poolSet := newRedisPoolSet(time.Second, 2, time.Second)
	defer poolSet.close()

	if pool, pool2 := poolSet.get(addr), poolSet.get(addr); pool2 != pool {
		t.Error("cannot reuse")
		t.Error(pool2)
		t.Fatal(pool)
	} else if pool3 := poolSet.get(addr2); pool3 == pool {
		t.Error("not new pool")
		t.Error(addr2)
		t.Fatal(addr)
	}
}
