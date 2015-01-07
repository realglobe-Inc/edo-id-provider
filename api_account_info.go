package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
)

const (
	scBear = "Bearer"
	scJws  = "JWS"
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

func accountInfoApi(sys *system, w http.ResponseWriter, r *http.Request) error {
	req, err := newAccountInfoRequest(r)
	if err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvReq, erro.Unwrap(err).Error())
	}

	reqTok := req.token()
	if reqTok == nil {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+headAuth)
	}

	log.Debug("Token entry is exist")

	tokId := reqTok.tokenId()
	if tokId == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no token")
	}

	log.Debug("Token " + mosaic(tokId) + " is declared")

	tok, err := sys.tokCont.get(tokId)
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is not exist")
	}

	log.Debug("Token " + mosaic(tokId) + " is exist")

	t, err := sys.taCont.get(tok.taId())
	if err != nil {
		return erro.Wrap(err)
	} else if t == nil {
		return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is linked to invalid TA "+tok.taId())
	}

	info := map[string]interface{}{}
	if reqTok.scheme() == scBear {
		log.Warn("Token type " + scBear + " is supported, but nonsense")
	} else {
		if err := reqTok.verify(t.keys()); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
			return responseError(w, http.StatusBadRequest, errInvTok, erro.Unwrap(err).Error())
		}
		acc, err := sys.accCont.get(tok.accountId())
		if err != nil {
			return erro.Wrap(err)
		} else if acc == nil {
			return responseError(w, http.StatusBadRequest, errInvTok, "token "+mosaic(tokId)+" is linked to invalid account "+tok.accountId())
		}
		for clmName := range tok.claims().Elements() {
			clm := acc.attribute(clmName)
			if clm == nil || clm == "" {
				continue
			}
			info[clmName] = clm
		}
	}

	return responseAccountInfo(w, info)
	// panic("not yet implemented")
}
