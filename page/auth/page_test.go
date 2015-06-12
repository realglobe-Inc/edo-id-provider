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

package auth

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"time"
)

func newTestPage(keys []jwk.Key, webs []webdb.Element, acnts []account.Element, tas []tadb.Element) *Page {
	return New(
		server.NewStopper(),
		test_idpId,
		"ES256",
		"",
		"/select.html",
		"/login.html",
		"/consent.html",
		nil,
		20,
		"Id-Provider",
		30,
		time.Minute,
		time.Minute/2,
		time.Hour,
		30,
		time.Minute,
		time.Hour,
		time.Minute,
		time.Minute,
		10,
		keydb.NewMemoryDb(keys),
		webdb.NewMemoryDb(webs),
		account.NewMemoryDb(acnts),
		consent.NewMemoryDb(),
		tadb.NewMemoryDb(tas),
		sector.NewMemoryDb(),
		pairwise.NewMemoryDb(),
		session.NewMemoryDb(),
		authcode.NewMemoryDb(),
		rand.New(time.Second),
		"/",
		false,
		true,
	)
}
