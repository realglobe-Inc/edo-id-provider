package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
)

const (
	scBear = "Bearer"
)

func responseAccountInfo(w http.ResponseWriter, info map[string]interface{}) error {
	buff, err := json.Marshal(info)
	if err != nil {
		return erro.Wrap(err)
	}

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
		return responseError(w, http.StatusBadRequest, errInvReq, "authorization scheme "+req.scheme()+" is not supported")
	}

	log.Debug("Authrization scheme " + req.scheme() + " is OK")

	tokId := req.token()
	if tokId == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no token")
	}

	log.Debug("Token " + mosaic(tokId) + " is declared")

	tok, err := sys.tokCont.get(tokId)
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is not exist")
	} else if !tok.valid() {
		return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is invalid")
	}

	log.Debug("Token " + mosaic(tokId) + " is exist")

	t, err := sys.taCont.get(tok.taId())
	if err != nil {
		return erro.Wrap(err)
	} else if t == nil {
		return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is linked to invalid TA "+tok.taId())
	}

	log.Debug("Token TA " + t.id() + " is exist")

	acc, err := sys.accCont.get(tok.accountId())
	if err != nil {
		return erro.Wrap(err)
	} else if acc == nil {
		return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is linked to invalid account "+tok.accountId())
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

	return responseAccountInfo(w, info)
}
