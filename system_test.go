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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"time"
)

func newTestSystem(selfKeys []jwk.Key, acnts []account.Element, tas []tadb.Element, idps []idpdb.Element, webs []webdb.Element) *system {
	return &system{
		//selfId:    "",
		sigAlg: "ES256",
		//sigKid:    "",
		pathTok:    test_pathTok,
		pathTa:     test_pathTa,
		pathSelUi:  test_pathSelUi,
		pathLginUi: test_pathLginUi,
		pathConsUi: test_pathConsUi,
		pathErrUi:  test_pathErrUi,

		pwSaltLen:    20,
		sessLabel:    "Id-Provider",
		sessLen:      30,
		sessExpIn:    time.Minute,
		sessRefDelay: time.Minute / 2,
		sessDbExpIn:  10 * time.Minute,
		acodLen:      30,
		acodExpIn:    time.Minute,
		acodDbExpIn:  10 * time.Minute,
		tokLen:       30,
		tokExpIn:     time.Minute,
		tokDbExpIn:   10 * time.Minute,
		ccodLen:      30,
		ccodExpIn:    time.Minute,
		ccodDbExpIn:  10 * time.Minute,
		jtiLen:       20,
		jtiExpIn:     time.Minute,
		jtiDbExpIn:   10 * time.Minute,
		ticLen:       10,

		keyDb:  keydb.NewMemoryDb(selfKeys),
		acntDb: account.NewMemoryDb(acnts),
		consDb: consent.NewMemoryDb(),
		taDb:   tadb.NewMemoryDb(tas),
		sectDb: sector.NewMemoryDb(),
		pwDb:   pairwise.NewMemoryDb(),
		idpDb:  idpdb.NewMemoryDb(idps),
		sessDb: session.NewMemoryDb(),
		acodDb: authcode.NewMemoryDb(),
		tokDb:  token.NewMemoryDb(),
		ccodDb: coopcode.NewMemoryDb(),
		jtiDb:  jtidb.NewMemoryDb(),
		webDb:  webdb.NewMemoryDb(webs),

		cookPath: "/",
		cookSec:  false,
	}
}
