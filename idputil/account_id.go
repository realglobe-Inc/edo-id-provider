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
	"github.com/realglobe-Inc/go-lib/erro"
)

type SetSubSystem interface {
	PairwiseSaltLength() int
	SectorDb() sector.Db
	PairwiseDb() pairwise.Db
	IdGenerator() rand.Generator
}

// ID トークンの sub クレームとして TA に通知するアカウント ID を設定する。
// 結果は acnt.SetAttribute にて登録される。
func SetSub(sys SetSubSystem, acnt account.Element, ta tadb.Element) error {
	if acnt.Attribute(tagSub) != nil {
		return nil
	} else if !ta.Pairwise() {
		acnt.SetAttribute(tagSub, acnt.Id())
		return nil
	}

	// セクタ固有のアカウント ID を計算。
	sect, err := sys.SectorDb().Get(ta.Sector())
	if err != nil {
		return erro.Wrap(err)
	} else if sect == nil {
		sect = sector.New(ta.Sector(), sys.IdGenerator().Bytes(sys.PairwiseSaltLength()))
		if existing, err := sys.SectorDb().SaveIfAbsent(sect); err != nil {
			return erro.Wrap(err)
		} else if existing != nil {
			sect = existing
		}
	}
	pw := pairwise.Generate(acnt.Id(), sect.Id(), sect.Salt())

	// TA 間連携で逆引きが必要になるので、セクタ固有のアカウント ID を保存。
	if err := sys.PairwiseDb().Save(pw); err != nil {
		return erro.Wrap(err)
	}

	acnt.SetAttribute(tagSub, pw.Pairwise())
	return nil
}
