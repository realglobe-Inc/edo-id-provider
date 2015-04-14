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
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"strconv"
	"time"
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

func tokenApi(w http.ResponseWriter, r *http.Request, sys *system) error {
	if r.Method != "POST" {
		return newIdpError(errInvReq, r.Method+" is not supported", http.StatusMethodNotAllowed, nil)
	}

	req := newTokenRequest(r)
	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return newIdpError(errInvReq, k+" is overlapped", http.StatusBadRequest, nil)
		}
	}

	if grntType := req.grantType(); grntType == "" {
		return newIdpError(errInvReq, "no "+formGrntType, http.StatusBadRequest, nil)
	} else if grntType != grntTypeCod {
		return newIdpError(errUnsuppGrntType, grntType+" is not supported", http.StatusBadRequest, nil)
	}

	log.Debug("Grant type is " + grntTypeCod)

	rawCod := req.code()
	if rawCod == "" {
		return newIdpError(errInvReq, "no "+formCod, http.StatusBadRequest, nil)
	}

	var codId string
	if codJt, err := jwt.Parse(rawCod, nil, nil); err != nil {
		// JWS から抜き出した ID だけ送られてきた。
		codId = rawCod
		rawCod = ""
	} else {
		// JWS のまま送られてきた。
		log.Debug("Raw code " + mosaic(rawCod) + " is declared")
		codId, _ = codJt.Claim(clmJti).(string)
	}

	log.Debug("Code " + mosaic(codId) + " is declared")

	cod, codTic, err := sys.codCont.getAndSetEntry(codId)
	if err != nil {
		return erro.Wrap(err)
	} else if cod == nil {
		return newIdpError(errInvGrnt, "code "+mosaic(codId)+" is not exist", http.StatusBadRequest, nil)
	} else if !cod.valid() {
		disposeCode(sys, codId)
		return newIdpError(errInvGrnt, "code "+mosaic(codId)+" is invalid", http.StatusBadRequest, nil)
	}

	log.Debug("Code " + mosaic(codId) + " is exist")

	// 認可コードを使用済みにする。
	cod.disable()

	log.Debug("Code " + mosaic(codId) + " was disabled")

	taId := req.taId()
	if taId == "" {
		return newIdpError(errInvReq, "no "+formTaId, http.StatusBadRequest, nil)
	} else if taId != cod.taId() {
		return newIdpError(errInvGrnt, "you are not code holder", http.StatusBadRequest, nil)
	} else {
		log.Debug("TA ID " + taId + " is declared")
	}

	rediUri := req.redirectUri()
	if rediUri == "" {
		return newIdpError(errInvReq, "no "+formRediUri, http.StatusBadRequest, nil)
	} else if rediUri != cod.redirectUri() {
		return newIdpError(errInvGrnt, "invalid "+formRediUri, http.StatusBadRequest, nil)
	}

	log.Debug(formRediUri + " matches that of code")

	if taAssType := req.taAssertionType(); taAssType == "" {
		return newIdpError(errInvTa, "no "+formTaAssType, http.StatusBadRequest, nil)
	} else if taAssType != taAssTypeJwt {
		return newIdpError(errInvTa, taAssType+" is not supported", http.StatusBadRequest, nil)
	}

	log.Debug(formTaAssType + " is " + taAssTypeJwt)

	taAss := req.taAssertion()
	if taAss == "" {
		return newIdpError(errInvTa, "no "+formTaAss, http.StatusBadRequest, nil)
	}

	log.Debug(formTaAss + " is found")

	// Authorization ヘッダと client_secret パラメータも認識はする。
	if r.Header.Get(headAuth) != "" || r.FormValue(formTaScrt) != "" {
		return newIdpError(errInvReq, "multi client authentication algorithms are exist", http.StatusBadRequest, nil)
	}

	// クライアント認証する。
	ta, err := sys.taCont.get(taId)
	if err != nil {
		return erro.Wrap(err)
	}

	assJt, err := jwt.Parse(taAss, ta.keys(), nil)
	if err != nil {
		err = erro.Wrap(err)
		return newIdpError(errInvTa, erro.Unwrap(err).Error(), http.StatusBadRequest, err)
	} else if assJt.Header(jwtAlg) == algNone {
		return newIdpError(errInvTa, "asserion "+jwtAlg+" must not be "+algNone, http.StatusBadRequest, nil)
	}

	now := time.Now()
	if assJt.Claim(clmIss) != taId {
		return newIdpError(errInvTa, "assertion "+clmIss+" is not "+taId, http.StatusBadRequest, nil)
	} else if assJt.Claim(clmSub) != taId {
		return newIdpError(errInvTa, "assertion "+clmSub+" is not "+taId, http.StatusBadRequest, nil)
	} else if jti := assJt.Claim(clmJti); jti == nil || jti == "" {
		return newIdpError(errInvTa, "no assertion "+clmJti, http.StatusBadRequest, nil)
	} else if exp, _ := assJt.Claim(clmExp).(float64); exp == 0 {
		return newIdpError(errInvTa, "no assertion "+clmExp, http.StatusBadRequest, nil)
	} else if intExp := int64(exp); exp != float64(intExp) {
		return newIdpError(errInvTa, "assertion "+clmExp+" is not integer", http.StatusBadRequest, nil)
	} else if intExp < now.Unix() {
		return newIdpError(errInvTa, "assertion expired", http.StatusBadRequest, nil)
	} else if aud := assJt.Claim(clmAud); aud == nil {
		return newIdpError(errInvTa, "no assertion "+clmAud, http.StatusBadRequest, nil)
	} else if !audienceHas(aud, sys.selfId+tokPath) {
		return newIdpError(errInvTa, "assertion "+clmAud+" does not contain "+sys.selfId+tokPath, http.StatusBadRequest, nil)
	} else if c := assJt.Claim(clmCod); !((rawCod != "" || c == rawCod) || c == codId) {
		return newIdpError(errInvTa, "invalid assertion "+clmCod, http.StatusBadRequest, nil)
	}

	// クライアント認証できた。
	log.Debug(taId + " is authenticated")

	tokId, err := sys.tokCont.newId()
	if err != nil {
		return erro.Wrap(err)
	}

	// アクセストークンが決まった。
	log.Debug("Token " + mosaic(tokId) + " was generated")

	idTokJt := jwt.New()
	idTokJt.SetHeader(jwtAlg, sys.sigAlg)
	if sys.sigKid != "" {
		idTokJt.SetHeader(jwtKid, sys.sigKid)
	}
	idTokJt.SetClaim(clmIss, sys.selfId)
	idTokJt.SetClaim(clmSub, cod.accountId())
	idTokJt.SetClaim(clmAud, cod.taId())
	idTokJt.SetClaim(clmExp, now.Add(sys.idTokExpiDur).Unix())
	idTokJt.SetClaim(clmIat, now.Unix())
	if !cod.authenticationDate().IsZero() {
		idTokJt.SetClaim(clmAuthTim, cod.authenticationDate().Unix())
	}
	if cod.nonce() != "" {
		idTokJt.SetClaim(clmNonc, cod.nonce())
	}
	if hGen, err := jwt.HashFunction(sys.sigAlg); err != nil {
		return erro.Wrap(err)
	} else if hGen > 0 {
		h := hGen.New()
		h.Write([]byte(tokId))
		sum := h.Sum(nil)
		idTokJt.SetClaim(clmAtHash, base64url.EncodeToString(sum[:len(sum)/2]))
	}
	buff, err := idTokJt.Encode(map[string]interface{}{sys.sigKid: sys.sigKey}, nil)
	if err != nil {
		return erro.Wrap(err)
	}
	idTok := string(buff)

	// ID トークンができた。
	log.Debug("ID token was generated")

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

	// アクセストークンを認可コードに結びつける。
	cod.addToken(tokId)
	if ok, err := sys.codCont.putIfEntered(cod, codTic); err != nil {
		return newIdpError(errServErr, erro.Unwrap(err).Error(), http.StatusBadRequest, erro.Wrap(err))
	} else if !ok {
		disposeCode(sys, codId)
		return newIdpError(errInvGrnt, "code "+mosaic(codId)+" is used by others", http.StatusBadRequest, nil)
	}

	log.Debug("Token " + mosaic(tok.id()) + " was linked to code " + mosaic(cod.id()))

	// アクセストークンを保存する。
	if err := sys.tokCont.put(tok); err != nil {
		return erro.Wrap(err)
	}

	log.Debug("Token " + mosaic(tok.id()) + " was registerd")

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

// 認可コードを廃棄処分する。
func disposeCode(sys *system, codId string) {
	cod, codTic, err := sys.codCont.getAndSetEntry(codId)
	if err != nil {
		// 何もできない。
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return
	} else if cod == nil {
		return
	}

	for tokId := range cod.tokens() {
		disposeToken(sys, tokId)
	}

	if !cod.valid() {
		return
	}

	cod.disable()
	if _, err := sys.codCont.putIfEntered(cod, codTic); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return
	}
}

// アクセストークンを廃棄処分する。
func disposeToken(sys *system, tokId string) {
	tok, err := sys.tokCont.get(tokId)
	if err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return
	} else if tok == nil {
		return
	} else if !tok.valid() {
		return
	}

	tok.disable()
	if err := sys.tokCont.put(tok); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return
	}
}
