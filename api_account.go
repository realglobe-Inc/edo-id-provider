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
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

func (sys *system) accountApi(w http.ResponseWriter, r *http.Request) error {
	req := newAccountRequest(r)

	if req.scheme() != scmBearer {
		return idperr.New(idperr.Invalid_request, "authorization scheme "+req.scheme()+" is not supported", http.StatusBadRequest, nil)
	}

	log.Debug("Authrization scheme " + req.scheme() + " is OK")

	if req.token() == "" {
		return idperr.New(idperr.Invalid_request, "no access token", http.StatusBadRequest, nil)
	}

	log.Debug("Access token " + mosaic(req.token()) + " is declared")

	tok, err := sys.tokDb.Get(req.token())
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return idperr.New(idperr.Invalid_token, "token "+mosaic(req.token())+" is not exist", http.StatusBadRequest, nil)
	} else if tok.Invalid() {
		return idperr.New(idperr.Invalid_token, "token "+mosaic(req.token())+" is invalid", http.StatusBadRequest, nil)
	}

	log.Debug("Declared access token " + mosaic(tok.Id()) + " is OK")

	ta, err := sys.taDb.Get(tok.Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if ta == nil {
		return idperr.New(idperr.Invalid_token, "TA "+tok.Ta()+" is not exist", http.StatusBadRequest, nil)
	}

	log.Debug("TA " + ta.Id() + " is exist")

	acnt, err := sys.acntDb.Get(tok.Account())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		return idperr.New(idperr.Invalid_token, "account is not exist", http.StatusBadRequest, nil)
	}

	log.Debug("Account " + acnt.Id() + " is exist")

	clms := scopeToClaims(tok.Scope())
	for clm := range tok.Attributes() {
		clms[clm] = true
	}

	log.Debug("Token claims ", clms, " will be returned")

	info := map[string]interface{}{}
	for clmName := range clms {
		clm := acnt.Attribute(clmName)
		if clm == nil || clm == "" {
			continue
		}
		info[clmName] = clm
	}
	info[clmSub] = acnt.Id()

	return response(w, info)
}
