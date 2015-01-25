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
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
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

	var codId string
	if codJws, err := util.ParseJws(rawCod); err != nil {
		// JWS から抜き出した ID だけ送られてきた。
		codId = rawCod
		rawCod = ""
	} else {
		// JWS のまま送られてきた。
		codId, _ = codJws.Claim(clmJti).(string)
	}

	log.Debug("Code " + mosaic(codId) + " is declared")

	cod, err := sys.codCont.get(codId)
	if err != nil {
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
	} else if cod == nil {
		return responseError(w, http.StatusBadRequest, errInvGrnt, "code "+mosaic(codId)+" is not exist")
	} else if !cod.valid() {
		// TODO 発行したアクセストークンを無効に。
		return responseError(w, http.StatusBadRequest, errInvGrnt, "code "+mosaic(codId)+" is invalid")
	}

	log.Debug("Code " + mosaic(codId) + " is exist")

	// 認可コードを使用済みにする。
	cod.disable()
	if err := sys.codCont.put(cod); err != nil {
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
	}

	log.Debug("Code " + mosaic(codId) + " is disabled")

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

	// クライアント認証する。
	assJws, err := util.ParseJws(taAss)
	if err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvTa, erro.Unwrap(err).Error())
	}

	now := time.Now()
	if assJws.Claim(clmIss) != taId {
		return responseError(w, http.StatusBadRequest, errInvTa, "assertion "+clmIss+" is not "+taId)
	} else if assJws.Claim(clmSub) != taId {
		return responseError(w, http.StatusBadRequest, errInvTa, "assertion "+clmSub+" is not "+taId)
	} else if jti := assJws.Claim(clmJti); jti == nil || jti == "" {
		return responseError(w, http.StatusBadRequest, errInvTa, "no assertion "+clmJti)
	} else if exp, _ := assJws.Claim(clmExp).(float64); exp == 0 {
		return responseError(w, http.StatusBadRequest, errInvTa, "no assertion "+clmExp)
	} else if intExp := int64(exp); exp != float64(intExp) {
		return responseError(w, http.StatusBadRequest, errInvTa, "assertion "+clmExp+" is not integer")
	} else if intExp < now.Unix() {
		return responseError(w, http.StatusBadRequest, errInvTa, "assertion expired")
	} else if aud := assJws.Claim(clmAud); aud == nil {
		return responseError(w, http.StatusBadRequest, errInvTa, "no assertion "+clmAud)
	} else if !audienceHas(aud, sys.selfId+tokPath) {
		return responseError(w, http.StatusBadRequest, errInvTa, "assertion "+clmAud+" does not contain "+sys.selfId+tokPath)
	} else if c := assJws.Claim(clmCod); !((rawCod != "" || c == rawCod) || c == codId) {
		return responseError(w, http.StatusBadRequest, errInvTa, "invalid assertion "+clmCod)
	}

	// クライアント認証情報は揃ってた。
	log.Debug("Assertion claims are OK")

	ta, err := sys.taCont.get(taId)
	if err != nil {
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
	}

	if assJws.Header(jwtAlg) == algNone {
		return responseError(w, http.StatusBadRequest, errInvTa, "asserion "+jwtAlg+" must not be "+algNone)
	} else if err := assJws.Verify(ta.keys()); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return responseError(w, http.StatusBadRequest, errInvTa, erro.Unwrap(err).Error())
	}

	// クライアント認証できた。
	log.Debug(taId + " is authenticated")

	idTokJws := util.NewJws()
	idTokJws.SetHeader(jwtAlg, sys.sigAlg)
	if sys.sigKid != "" {
		idTokJws.SetHeader(jwtKid, sys.sigKid)
	}
	idTokJws.SetClaim(clmIss, sys.selfId)
	idTokJws.SetClaim(clmSub, cod.accountId())
	idTokJws.SetClaim(clmAud, cod.taId())
	idTokJws.SetClaim(clmExp, now.Add(sys.idTokExpiDur).Unix())
	idTokJws.SetClaim(clmIat, now.Unix())
	if !cod.authenticationDate().IsZero() {
		idTokJws.SetClaim(clmAuthTim, cod.authenticationDate().Unix())
	}
	if cod.nonce() != "" {
		idTokJws.SetClaim(clmNonc, cod.nonce())
	}
	if err := idTokJws.Sign(map[string]crypto.PrivateKey{sys.sigKid: sys.sigKey}); err != nil {
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
	}
	buff, err := idTokJws.Encode()
	if err != nil {
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
	}
	idTok := string(buff)

	// ID トークンができた。
	log.Debug("ID token was generated")

	tokId, err := sys.tokCont.newId()
	if err != nil {
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
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
		return responseServerError(w, http.StatusBadRequest, erro.Wrap(err))
	}

	// アクセストークンが決まった。
	log.Debug("Token " + mosaic(tok.id()) + " is generated")

	return responseToken(w, tok)
}

// aud クレーム値が tgt を含むかどうか検査。
func audienceHas(aud interface{}, tgt string) bool {
	switch a := aud.(type) {
	case string:
		return a == tgt
	case []interface{}:
		for _, elem := range a {
			s, _ := elem.(string)
			if s == tgt {
				return true
			}
		}
		return false
	default:
		return false
	}
}
