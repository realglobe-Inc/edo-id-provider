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

package idputil

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/rand"
	"testing"
	"time"
)

type setSubSystem struct {
	pwSaltLen int
	sectDb    sector.Db
	pwDb      pairwise.Db
	idGen     rand.Generator
}

func (this *setSubSystem) PairwiseSaltLength() int     { return this.pwSaltLen }
func (this *setSubSystem) SectorDb() sector.Db         { return this.sectDb }
func (this *setSubSystem) PairwiseDb() pairwise.Db     { return this.pwDb }
func (this *setSubSystem) IdGenerator() rand.Generator { return this.idGen }

func newSetSubSystem() SetSubSystem {
	return &setSubSystem{
		20,
		sector.NewMemoryDb(),
		pairwise.NewMemoryDb(),
		rand.New(time.Millisecond),
	}
}

// 共通 ID の場合。
func TestSetSubPublic(t *testing.T) {
	sys := newSetSubSystem()
	acnt := account.New(test_acntId, test_acntName, nil, nil)
	ta := tadb.New(test_taId, nil, nil, nil, false, test_sectId)
	if err := SetSub(sys, acnt, ta); err != nil {
		t.Fatal(err)
	} else if sub, _ := acnt.Attribute("sub").(string); sub != test_acntId {
		t.Error(sub)
		t.Fatal(test_acntId)
	}
}

// TA 固有 ID の場合。
func TestSetSubPairwise(t *testing.T) {
	sys := newSetSubSystem()
	acnt := account.New(test_acntId, test_acntName, nil, nil)
	ta := tadb.New(test_taId, nil, nil, nil, true, test_sectId)
	if err := SetSub(sys, acnt, ta); err != nil {
		t.Fatal(err)
	} else if sub, _ := acnt.Attribute("sub").(string); sub == "" || sub == test_acntId {
		t.Error(sub)
		t.Fatal(test_acntId)
	}
}
