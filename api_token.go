package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"time"
)

const (
	jwtAlg = "alg"
	jwtKid = "kid"
)

const (
	algNone = "none"
)

func responseToken(w http.ResponseWriter, tok *token) error {
	m := map[string]interface{}{
		formTokId:   tok.Id,
		formTokType: tokTypeBear,
	}
	if !tok.ExpiDate.IsZero() {
		m[formExpi] = int64(tok.ExpiDate.Sub(time.Now()).Seconds())
	}
	if tok.RefTok != "" {
		m[formRefTok] = tok.RefTok
	}
	if len(tok.Scops) > 0 {
		var buff string
		for scop := range tok.Scops {
			if len(buff) > 0 {
				buff += " "
			}
			buff += scop
		}
		m[formScop] = buff
	}
	buff, err := json.Marshal(m)
	if err != nil {
		return erro.Wrap(err)
	}

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	if _, err := w.Write(buff); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
	}
	return nil
}

func tokenApi(sys *system, w http.ResponseWriter, r *http.Request) error {
	req := newTokenRequest(r)

	if grntType := req.grantType(); grntType == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+formGrntType)
	} else if grntType == grntTypeCod {
		return responseError(w, http.StatusBadRequest, errUnsuppGrntType, grntType+" is not supported")
	}

	log.Debug("Grant type is " + grntTypeCod)

	codId := req.code()
	if codId == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+formCod)
	}

	log.Debug("Code " + mosaic(codId) + " is declared")

	cod, err := sys.codCont.get(codId)
	if err != nil {
		return erro.Wrap(err)
	} else if cod == nil {
		return responseError(w, http.StatusBadRequest, errInvGrnt, "code "+codId+" is not exist")
	}

	log.Debug("Code " + mosaic(codId) + " is exist")

	taId := req.taId()
	if taId == "" {
		taId = cod.TaId
		log.Debug("TA ID is " + taId + " in code")
	} else if taId != cod.TaId {
		return responseError(w, http.StatusBadRequest, errInvTa, "you are not code holder")
	} else {
		log.Debug("TA ID " + taId + " is declared")
	}

	rediUri := req.redirectUri()
	if rediUri == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+formRediUri)
	} else if rediUri != cod.RediUri {
		return responseError(w, http.StatusBadRequest, errInvTa, "invalid "+formRediUri)
	}

	log.Debug(formRediUri + " matches that of code")

	if taAssType := req.taAssertionType(); taAssType == "" {
		return responseError(w, http.StatusBadRequest, errInvTa, "no "+formTaAssType)
	} else if taAssType != taAssTypeJwt {
		return responseError(w, http.StatusBadRequest, errInvTa, taAssType+" is not supported")
	}

	log.Debug(formTaAssType + " is " + taAssTypeJwt)

	taAss := req.taAssertion()
	if taAss == "" {
		return responseError(w, http.StatusBadRequest, errInvTa, "no "+formTaAss)
	}

	log.Debug(formTaAss + " is found")

	jws, err := util.ParseJws(taAss)
	if err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvTa, erro.Unwrap(err).Error())
	}

	if jws.Claim(clmIss) != taId {
		return responseError(w, http.StatusBadRequest, errInvTa, clmIss+" is not "+taId)
	} else if jws.Claim(clmSub) != taId {
		return responseError(w, http.StatusBadRequest, errInvTa, clmSub+" is not "+taId)
	} else if jti := jws.Claim(clmJti); jti == nil || jti == "" {
		return responseError(w, http.StatusBadRequest, errInvTa, "no "+clmJti)
	}
	exp, _ := jws.Claim(clmExp).(float64)
	if exp == 0 {
		return responseError(w, http.StatusBadRequest, errInvTa, "no "+clmExp)
	} else if exp != float64(int64(exp)) {
		return responseError(w, http.StatusBadRequest, errInvTa, clmExp+" is not integer")
	}
	aud := jws.Claim(clmAud)
	if aud == nil {
		return responseError(w, http.StatusBadRequest, errInvTa, "no "+clmAud)
	}
	switch a := aud.(type) {
	case string:
		if a != sys.selfId+tokPath {
			return responseError(w, http.StatusBadRequest, errInvTa, clmAud+" is not "+sys.selfId+tokPath)
		}
	case []interface{}:
		ok := false
		for _, b := range a {
			c, _ := b.(string)
			if c == sys.selfId {
				ok = true
				break
			}
		}
		if !ok {
			return responseError(w, http.StatusBadRequest, errInvTa, clmAud+" does not have "+sys.selfId)
		}
	default:
		return responseError(w, http.StatusBadRequest, errInvTa, "invalid"+clmAud)
	}
	if c := jws.Claim(clmCod); c != codId {
		return responseError(w, http.StatusBadRequest, errInvTa, "invalid"+clmCod)
	}

	log.Debug("JWS claims are OK")

	ta, err := sys.taCont.get(taId)
	if err != nil {
		return erro.Wrap(err)
	}

	if alg := jws.Header(jwtAlg); alg == algNone {
		return responseError(w, http.StatusBadRequest, errInvTa, "JWS algorithm "+algNone+" is not allowed")
	} else if err := jws.Verify(ta.pubKeys); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvTa, erro.Unwrap(err).Error())
	}

	log.Debug(taId + " is authenticated")

	tok, err := sys.tokCont.new(cod.AccId, 0)
	if err != nil {
		return erro.Wrap(err)
	}

	log.Debug("Token " + mosaic(tok.Id) + " is generated")

	return responseToken(w, tok)
}
