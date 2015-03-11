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
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"strconv"
)

func responseAccountInfo(w http.ResponseWriter, info map[string]interface{}) error {
	buff, err := json.Marshal(info)
	if err != nil {
		return newIdpError(errServErr, erro.Unwrap(err).Error(), http.StatusBadRequest, erro.Wrap(err))
	}

	w.Header().Add("Content-Type", server.ContentTypeJson)
	w.Header().Add("Content-Length", strconv.Itoa(len(buff)))
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	if _, err := w.Write(buff); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
	}
	return nil
}

func accountInfoApi(w http.ResponseWriter, r *http.Request, sys *system) error {
	req := newAccountInfoRequest(r)

	if req.scheme() != scBear {
		return newIdpError(errInvReq, "authorization scheme "+req.scheme()+" is not supported", http.StatusBadRequest, nil)
	}

	log.Debug("Authrization scheme " + req.scheme() + " is OK")

	tokId := req.token()
	if tokId == "" {
		return newIdpError(errInvReq, "no token", http.StatusBadRequest, nil)
	}

	log.Debug("Token " + mosaic(tokId) + " is declared")

	tok, err := sys.tokCont.get(tokId)
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return newIdpError(errInvTok, "token "+mosaic(tokId)+" is not exist", http.StatusBadRequest, nil)
	} else if !tok.valid() {
		return newIdpError(errInvTok, "token "+mosaic(tokId)+" is invalid", http.StatusBadRequest, nil)
	}

	log.Debug("Token " + mosaic(tokId) + " is exist")

	t, err := sys.taCont.get(tok.taId())
	if err != nil {
		return erro.Wrap(err)
	} else if t == nil {
		return newIdpError(errInvTok, "token "+mosaic(tokId)+" is linked to invalid TA "+tok.taId(), http.StatusBadRequest, nil)
	}

	log.Debug("Token TA " + t.id() + " is exist")

	acc, err := sys.accCont.get(tok.accountId())
	if err != nil {
		return erro.Wrap(err)
	} else if acc == nil {
		return newIdpError(errInvTok, "token "+mosaic(tokId)+" is linked to invalid account "+tok.accountId(), http.StatusBadRequest, nil)
	}

	log.Debug("Token account " + acc.id() + " is exist")

	clms := scopesToClaims(tok.scopes())
	for clm := range tok.claims() {
		clms[clm] = true
	}

	log.Debug("Token claims ", clms, " will be returned")

	info := map[string]interface{}{}
	for clmName := range clms {
		clm := acc.attribute(clmName)
		if clm == nil || clm == "" {
			continue
		}
		info[clmName] = clm
	}
	info[clmSub] = acc.id()

	return responseAccountInfo(w, info)
}
