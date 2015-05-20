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
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

func (sys *system) accountApi(w http.ResponseWriter, r *http.Request) error {
	sender := request.Parse(r, "")

	req := newAccountRequest(r)

	if req.scheme() != tagBearer {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported authorization scheme "+req.scheme(), http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Authrization scheme "+req.scheme()+" is OK")

	if req.token() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no access token", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Access token "+mosaic(req.token())+" is declared")

	tok, err := sys.tokDb.Get(req.token())
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "token is not exist", http.StatusBadRequest, nil))
	} else if tok.Invalid() {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "token is invalid", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Declared access token "+mosaic(tok.Id())+" is OK")

	ta, err := sys.taDb.Get(tok.Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if ta == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "TA "+tok.Ta()+" is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": TA "+ta.Id()+" is exist")

	acnt, err := sys.acntDb.Get(tok.Account())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_token, "account is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Account "+acnt.Id()+" is exist")

	if err := sys.setSub(acnt, ta); err != nil {
		return erro.Wrap(err)
	}

	clms := scopeToClaims(tok.Scope())
	for clm := range tok.Attributes() {
		clms[clm] = true
	}
	clms[tagSub] = true

	log.Debug(sender, ": Return claims ", clms)

	info := map[string]interface{}{}
	for clmName := range clms {
		clm := acnt.Attribute(clmName)
		if clm == nil || clm == "" {
			continue
		}
		info[clmName] = clm
	}

	return response(w, info)
}
