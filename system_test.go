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
	acntapi "github.com/realglobe-Inc/edo-id-provider/api/account"
	"github.com/realglobe-Inc/edo-id-provider/api/coopfrom"
	"github.com/realglobe-Inc/edo-id-provider/api/coopto"
	tokapi "github.com/realglobe-Inc/edo-id-provider/api/token"
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
	authpage "github.com/realglobe-Inc/edo-id-provider/page/auth"
	taapi "github.com/realglobe-Inc/edo-idp-selector/api/ta"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"html/template"
	"net/http"
	"time"
)

type system struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	pathTok    string
	pathCoopFr string
	pathCoopTo string
	pathTa     string
	pathSelUi  string
	pathLginUi string
	pathConsUi string

	errTmpl *template.Template

	pwSaltLen    int
	sessLabel    string
	sessLen      int
	sessExpIn    time.Duration
	sessRefDelay time.Duration
	sessDbExpIn  time.Duration
	acodLen      int
	acodExpIn    time.Duration
	acodDbExpIn  time.Duration
	tokLen       int
	tokExpIn     time.Duration
	tokDbExpIn   time.Duration
	ccodLen      int
	ccodExpIn    time.Duration
	ccodDbExpIn  time.Duration
	jtiLen       int
	jtiExpIn     time.Duration
	jtiDbExpIn   time.Duration
	ticLen       int
	ticExpIn     time.Duration

	keyDb  keydb.Db
	webDb  webdb.Db
	acntDb account.Db
	consDb consent.Db
	taDb   tadb.Db
	sectDb sector.Db
	pwDb   pairwise.Db
	idpDb  idpdb.Db
	sessDb session.Db
	acodDb authcode.Db
	tokDb  token.Db
	ccodDb coopcode.Db
	jtiDb  jtidb.Db
	idGen  rand.Generator

	cookPath string
	cookSec  bool
	debug    bool
}

func newTestSystem(selfKeys []jwk.Key, acnts []account.Element, tas []tadb.Element, idps []idpdb.Element, webs []webdb.Element) *system {
	return &system{
		server.NewStopper(),
		"",
		"ES256",
		"",
		test_pathTok,
		test_pathCoopFr,
		test_pathCoopTo,
		test_pathTa,
		test_pathSelUi,
		test_pathLginUi,
		test_pathConsUi,
		nil,
		20,
		"Id-Provider",
		30,
		time.Minute,
		time.Minute / 2,
		10 * time.Minute,
		30,
		time.Minute,
		10 * time.Minute,
		30,
		time.Minute,
		10 * time.Minute,
		30,
		time.Minute,
		10 * time.Minute,
		20,
		time.Minute,
		10 * time.Minute,
		10,
		time.Minute,
		keydb.NewMemoryDb(selfKeys),
		webdb.NewMemoryDb(webs),
		account.NewMemoryDb(acnts),
		consent.NewMemoryDb(),
		tadb.NewMemoryDb(tas),
		sector.NewMemoryDb(),
		pairwise.NewMemoryDb(),
		idpdb.NewMemoryDb(idps),
		session.NewMemoryDb(),
		authcode.NewMemoryDb(),
		token.NewMemoryDb(),
		coopcode.NewMemoryDb(),
		jtidb.NewMemoryDb(),
		rand.New(time.Millisecond),
		"/",
		false,
		true,
	}
}

func (this *system) authPage() *authpage.Page {
	return authpage.New(
		this.stopper,
		this.selfId,
		this.sigAlg,
		this.sigKid,
		this.pathSelUi,
		this.pathLginUi,
		this.pathConsUi,
		this.errTmpl,
		this.pwSaltLen,
		this.sessLabel,
		this.sessLen,
		this.sessExpIn,
		this.sessRefDelay,
		this.sessDbExpIn,
		this.acodLen,
		this.acodExpIn,
		this.acodDbExpIn,
		this.tokExpIn,
		this.jtiExpIn,
		this.ticLen,
		this.ticExpIn,
		this.keyDb,
		this.webDb,
		this.acntDb,
		this.consDb,
		this.taDb,
		this.sectDb,
		this.pwDb,
		this.sessDb,
		this.acodDb,
		this.idGen,
		this.cookPath,
		this.cookSec,
		this.debug,
	)
}

func (this *system) taApi() http.Handler {
	return taapi.New(
		this.stopper,
		this.pathTa,
		this.taDb,
		this.debug,
	)
}

func (this *system) tokenApi() http.Handler {
	return tokapi.New(
		this.stopper,
		this.selfId,
		this.sigAlg,
		this.sigKid,
		this.pathTok,
		this.pwSaltLen,
		this.tokLen,
		this.tokExpIn,
		this.tokDbExpIn,
		this.jtiExpIn,
		this.keyDb,
		this.acntDb,
		this.taDb,
		this.sectDb,
		this.pwDb,
		this.acodDb,
		this.tokDb,
		this.jtiDb,
		this.idGen,
		this.debug,
	)
}

func (this *system) accountApi() http.Handler {
	return acntapi.New(
		this.stopper,
		this.pwSaltLen,
		this.acntDb,
		this.taDb,
		this.sectDb,
		this.pwDb,
		this.tokDb,
		this.idGen,
		this.debug,
	)
}

func (this *system) coopFromApi() http.Handler {
	return coopfrom.New(
		this.stopper,
		this.selfId,
		this.sigAlg,
		this.sigKid,
		this.pathCoopFr,
		this.ccodLen,
		this.ccodExpIn,
		this.ccodDbExpIn,
		this.jtiLen,
		this.jtiExpIn,
		this.keyDb,
		this.pwDb,
		this.acntDb,
		this.taDb,
		this.idpDb,
		this.ccodDb,
		this.tokDb,
		this.jtiDb,
		this.idGen,
		this.debug,
	)
}

func (this *system) coopToApi() http.Handler {
	return coopto.New(
		this.stopper,
		this.selfId,
		this.sigAlg,
		this.sigKid,
		this.pathCoopTo,
		this.pwSaltLen,
		this.tokLen,
		this.tokExpIn,
		this.tokDbExpIn,
		this.jtiExpIn,
		this.keyDb,
		this.acntDb,
		this.consDb,
		this.taDb,
		this.sectDb,
		this.pwDb,
		this.ccodDb,
		this.tokDb,
		this.jtiDb,
		this.idGen,
		this.debug,
	)
}
