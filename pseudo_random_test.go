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
	"encoding/base64"
	"testing"
	"time"
)

func TestPseudoRandomString(t *testing.T) {
	prand := newPseudoRandom(time.Millisecond)

	m := map[string]bool{}
	for j := 0; j < 100; j++ {
		for i := 100; i < 200; i++ {
			id := prand.string(i)
			if m[id] {
				t.Fatal("overlap " + id)
			} else if len(id) != i {
				t.Error(id)
				t.Error(len(id))
				t.Fatal(i)
			}
			m[id] = true
		}
	}
}

func TestPseudoRandomBytes(t *testing.T) {
	prand := newPseudoRandom(time.Millisecond)

	m := map[string]bool{}
	for j := 0; j < 100; j++ {
		for i := 100; i < 200; i++ {
			id := prand.bytes(i)
			label := base64.StdEncoding.EncodeToString(id)
			if m[label] {
				t.Fatal("overlap " + label)
			} else if len(id) != i {
				t.Error(label)
				t.Error(len(id))
				t.Fatal(i)
			}
		}
	}
}
