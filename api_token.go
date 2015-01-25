package main

import (
	"crypto"
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

const (
	grntTypeCod = "code"
)

func responseToken(w http.ResponseWriter, tok *token) error {
	m := map[string]interface{}{
		formTokId:   tok.id(),
		formTokType: tokTypeBear,
	}
	if !tok.expirationDate().IsZero() {
		m[formExpi] = int64(tok.expirationDate().Sub(time.Now()).Seconds())
	}
	if tok.refreshToken() != "" {
		m[formRefTok] = tok.refreshToken()
	}
	if len(tok.scopes()) > 0 {
		var buff string
		for scop := range tok.scopes() {
			if len(buff) > 0 {
				buff += " "
			}
			buff += scop
		}
		m[formScop] = buff
	}
	if tok.idToken() != "" {
		m[formIdTok] = tok.idToken()
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

func tokenApi(w http.ResponseWriter, r *http.Request, sys *system) error {
	req := newTokenRequest(r)

	if grntType := req.grantType(); grntType == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+formGrntType)
	} else if grntType == grntTypeCod {
		return responseError(w, http.StatusBadRequest, errUnsuppGrntType, grntType+" is not supported")
	}

	log.Debug("Grant type is " + grntTypeCod)

	rawCod := req.code()
	if rawCod == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+formCod)
	}

	log.Debug("Raw code " + mosaic(rawCod) + " is declared")

	// 認可コードを JWS として解釈。
	jws, err := util.ParseJws(rawCod)
	if err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvReq, erro.Unwrap(err).Error())
	}
	codId, _ := jws.Claim(clmJti).(string)

	log.Debug("Code " + mosaic(codId) + " is declared")

	cod, err := sys.codCont.get(codId)
	if err != nil {
		return erro.Wrap(err)
	} else if cod == nil {
		return responseError(w, http.StatusBadRequest, errInvGrnt, "code "+mosaic(codId)+" is not exist")
	} else if !cod.valid() {
		// TODO 発行したアクセストークンを無効に。
		return responseError(w, http.StatusBadRequest, errInvGrnt, "code "+mosaic(codId)+" is invalid")
	}

	log.Debug("Code " + mosaic(codId) + " is exist")

	taId := req.taId()
	if taId == "" {
		taId = cod.taId()
		log.Debug("TA ID is " + taId + " in code")
	} else if taId != cod.taId() {
		return responseError(w, http.StatusBadRequest, errInvTa, "you are not code holder")
	} else {
		log.Debug("TA ID " + taId + " is declared")
	}

	rediUri := req.redirectUri()
	if rediUri == "" {
		return responseError(w, http.StatusBadRequest, errInvReq, "no "+formRediUri)
	} else if rediUri != cod.redirectUri() {
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

	jws, err = util.ParseJws(taAss)
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
			if c == sys.selfId+tokPath {
				ok = true
				break
			}
		}
		if !ok {
			return responseError(w, http.StatusBadRequest, errInvTa, clmAud+" does not have "+sys.selfId+tokPath)
		}
	default:
		return responseError(w, http.StatusBadRequest, errInvTa, "invalid "+clmAud)
	}
	if c := jws.Claim(clmCod); c != rawCod {
		return responseError(w, http.StatusBadRequest, errInvTa, "invalid "+clmCod)
	}

	log.Debug("JWS claims are OK")

	ta, err := sys.taCont.get(taId)
	if err != nil {
		return erro.Wrap(err)
	}

	if jws.Header(jwtAlg) == algNone {
		return responseError(w, http.StatusBadRequest, errInvTa, "JWS algorithm "+algNone+" is not allowed")
	} else if err := jws.Verify(ta.keys()); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvTa, erro.Unwrap(err).Error())
	}

	log.Debug(taId + " is authenticated")

	jws = util.NewJws()
	jws.SetHeader(jwtAlg, sys.sigAlg)
	if sys.sigKid != "" {
		jws.SetHeader(jwtKid, sys.sigKid)
	}
	jws.SetClaim(clmIss, sys.selfId)
	jws.SetClaim(clmSub, cod.accountId())
	jws.SetClaim(clmAud, cod.taId())
	now := time.Now()
	jws.SetClaim(clmExp, now.Add(sys.idTokExpiDur).Unix())
	jws.SetClaim(clmIat, now.Unix())
	if !cod.authenticationDate().IsZero() {
		jws.SetClaim(clmAuthTim, cod.authenticationDate().Unix())
	}
	if cod.nonce() != "" {
		jws.SetClaim(clmNonc, cod.nonce())
	}
	if err := jws.Sign(map[string]crypto.PrivateKey{sys.sigKid: sys.sigKey}); err != nil {
		return erro.Wrap(err)
	}
	buff, err := jws.Encode()
	if err != nil {
		return erro.Wrap(err)
	}
	idTok := string(buff)

	// ID トークンができた。
	log.Debug("ID token was generated")

	tokId, err := sys.tokCont.newId()
	if err != nil {
		return erro.Wrap(err)
	}
	tok := newToken(
		tokId,
		cod.accountId(),
		cod.taId(),
		cod.id(),
		"",
		now.Add(cod.expirationDuration()),
		cod.scopes(),
		cod.claims(),
		idTok,
	)
	if err := sys.tokCont.put(tok); err != nil {
		return erro.Wrap(err)
	}

	// アクセストークンが決まった。
	log.Debug("Token " + mosaic(tok.id()) + " is generated")

	return responseToken(w, tok)
}
