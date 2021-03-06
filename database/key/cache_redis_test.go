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

package key

import (
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/test"
)

const (
	test_tag = "edo-test"
)

func TestRedisCache(t *testing.T) {
	red, err := test.NewRedisServer()
	if err != nil {
		t.Fatal(err)
	} else if red == nil {
		t.SkipNow()
	}
	defer red.Close()

	testCache(t, NewRedisCache(NewMemoryDb([]jwk.Key{test_key, test_sigKey, test_encKey, test_veriKey}), red.Pool(), test_tag, time.Minute))
}
