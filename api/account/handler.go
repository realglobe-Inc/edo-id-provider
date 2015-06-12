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

// アカウント情報エンドポイント。
package account

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	"github.com/realglobe-Inc/edo-id-provider/idputil"
	"github.com/realglobe-Inc/edo-id-provider/scope"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
)

type handler struct {
	stopper *server.Stopper

	pwSaltLen int

	acntDb account.Db
	taDb   tadb.Db
	sectDb sector.Db
	pwDb   pairwise.Db
	tokDb  token.Db
	idGen  rand.Generator

	debug bool
}

func New(
	stopper *server.Stopper,
	pwSaltLen int,
	acntDb account.Db,
	taDb tadb.Db,
	sectDb sector.Db,
	pwDb pairwise.Db,
	tokDb token.Db,
	idGen rand.Generator,
	debug bool,
) http.Handler {
	return &handler{
		stopper:   stopper,
		pwSaltLen: pwSaltLen,
		acntDb:    acntDb,
		taDb:      taDb,
		sectDb:    sectDb,
		pwDb:      pwDb,
		tokDb:     tokDb,
		idGen:     idGen,
		debug:     debug,
	}
}

func (this *handler) PairwiseSaltLength() int     { return this.pwSaltLen }
func (this *handler) SectorDb() sector.Db         { return this.sectDb }
func (this *handler) PairwiseDb() pairwise.Db     { return this.pwDb }
func (this *handler) IdGenerator() rand.Generator { return this.idGen }

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sender *requtil.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondJson(w, r, erro.New(rcv), sender)
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, this.debug)
	//////////////////////////////

	sender = requtil.Parse(r, "")
	log.Info(sender, ": Received account request")
	defer log.Info(sender, ": Handled account request")

	if err := this.serve(w, r, sender); err != nil {
		idperr.RespondJson(w, r, erro.Wrap(err), sender)
		return
	}
}

func (this *handler) serve(w http.ResponseWriter, r *http.Request, sender *requtil.Request) error {
	req, err := parseRequest(r)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if req.scheme() != tagBearer {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported authorization scheme "+req.scheme(), http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Authrization scheme "+req.scheme()+" is OK")

	if req.token() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no access token", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Access token "+logutil.Mosaic(req.token())+" is declared")

	tok, err := this.tokDb.Get(req.token())
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "token is not exist", http.StatusBadRequest, nil))
	} else if tok.Invalid() {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "token is invalid", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Declared access token "+logutil.Mosaic(tok.Id())+" is OK")

	ta, err := this.taDb.Get(tok.Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if ta == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "TA "+tok.Ta()+" is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": TA "+ta.Id()+" is exist")

	acnt, err := this.acntDb.Get(tok.Account())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "account is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Account "+acnt.Id()+" is exist")

	if err := idputil.SetSub(this, acnt, ta); err != nil {
		return erro.Wrap(err)
	}

	attrNames := map[string]bool{}
	for scop := range tok.Scope() {
		for attrName := range scope.Attributes(scop) {
			attrNames[attrName] = true
		}
	}
	for attrName := range tok.Attributes() {
		attrNames[attrName] = true
	}
	attrNames[tagSub] = true

	log.Debug(sender, ": Return attributes ", attrNames)

	attrs := map[string]interface{}{}
	for attrName := range attrNames {
		attr := acnt.Attribute(attrName)
		if attr == nil || attr == "" {
			continue
		}
		attrs[attrName] = attr
	}

	return idputil.RespondJson(w, attrs)
}
