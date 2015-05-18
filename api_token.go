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
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	"github.com/realglobe-Inc/edo-id-provider/request"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

func responseToken(w http.ResponseWriter, tok *token.Element, refTok, idTok string) error {
	m := map[string]interface{}{
		formAccess_token: tok.Id(),
		formToken_type:   tokTypeBearer,
	}
	if !tok.Expires().IsZero() {
		m[formExpires_in] = int64(tok.Expires().Sub(time.Now()).Seconds())
	}
	if len(tok.Scope()) > 0 {
		var buff string
		for scop := range tok.Scope() {
			if len(buff) > 0 {
				buff += " "
			}
			buff += scop
		}
		m[formScope] = buff
	}
	if refTok != "" {
		m[formRefresh_token] = refTok
	}
	if idTok != "" {
		m[formId_token] = idTok
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
		log.Err(erro.Wrap(err))
	}
	return nil
}

func (sys *system) tokenApi(w http.ResponseWriter, r *http.Request) error {
	sender := request.Parse(r, "")

	if r.Method != "POST" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported method "+r.Method, http.StatusMethodNotAllowed, nil))
	}

	req := newTokenRequest(r)
	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "parameter "+k+" overlaps", http.StatusBadRequest, nil))
		}
	}

	if grntType := req.grantType(); grntType == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+formGrant_type, http.StatusBadRequest, nil))
	} else if grntType != grntTypeAuthorization_code {
		return erro.Wrap(idperr.New(idperr.Unsupported_grant_type, "unsupported grant type "+grntType, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Grant type is "+grntTypeAuthorization_code)

	codId := req.code()
	if codId == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+formCode, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Code "+mosaic(codId)+" is declared")

	now := time.Now()

	cod, err := sys.acodDb.Get(codId)
	if err != nil {
		return erro.Wrap(err)
	} else if cod == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+mosaic(codId)+" is not exist", http.StatusBadRequest, nil))
	} else if cod.Expires().Before(now) {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+mosaic(codId)+" is expired", http.StatusBadRequest, nil))
	} else if cod.Token() != "" {
		disposeCode(sys, codId)
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+mosaic(codId)+" is invalid", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Code "+mosaic(codId)+" is exist")
	savedCodDate := cod.Date()

	if req.ta() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+formClient_id, http.StatusBadRequest, nil))
	} else if req.ta() != cod.Ta() {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "you are not code holder", http.StatusBadRequest, nil))
	} else {
		log.Debug(sender, ": TA "+req.ta()+" is declared")
	}

	rediUri := req.redirectUri()
	if rediUri == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+formRedirect_uri, http.StatusBadRequest, nil))
	} else if !reflect.DeepEqual(rediUri, cod.RedirectUri()) {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "invalid "+formRedirect_uri, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+formRedirect_uri+" matches that of code")

	if taAssType := req.taAssertionType(); taAssType == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no "+formClient_assertion_type, http.StatusBadRequest, nil))
	} else if taAssType != taAssTypeJwt {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "unsupported assertion type "+taAssType, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+formClient_assertion_type+" is "+taAssTypeJwt)

	taAss := req.taAssertion()
	if taAss == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no "+formClient_assertion, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+formClient_assertion+" is found")

	// Authorization ヘッダと client_secret パラメータも認識はする。
	if r.Header.Get(headAuthorization) != "" || r.FormValue(formClient_secret) != "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "multi client authentication algorithms are exist", http.StatusBadRequest, nil))
	}

	// クライアント認証する。
	ta, err := sys.taDb.Get(req.ta())
	if err != nil {
		return erro.Wrap(err)
	}

	assJt, err := jwt.Parse(taAss)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if assJt.Header(jwtAlg) == algNone {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "asserion "+jwtAlg+" must not be "+algNone, http.StatusBadRequest, nil))
	} else if err := assJt.Verify(ta.Keys()); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	}

	if assJt.Claim(clmIss) != req.ta() {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "JWT issuer is not "+req.ta(), http.StatusBadRequest, nil))
	} else if jti, _ := assJt.Claim(clmJti).(string); jti == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no JWT ID", http.StatusBadRequest, nil))
	} else if rawExp, _ := assJt.Claim(clmExp).(float64); rawExp == 0 {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no expiration date", http.StatusBadRequest, nil))
	} else if exp := time.Unix(int64(rawExp), 0); now.After(exp) {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "assertion expired", http.StatusBadRequest, nil))
	} else if ok, err := sys.jtiDb.SaveIfAbsent(jtidb.New(req.ta(), jti, exp)); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "overlapped JWT ID", http.StatusBadRequest, nil))
	} else if assJt.Claim(clmSub) != req.ta() {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "JWT subject is not "+req.ta(), http.StatusBadRequest, nil))
	} else if aud := assJt.Claim(clmAud); aud == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no assertion "+clmAud, http.StatusBadRequest, nil))
	} else if !audienceHas(aud, sys.selfId+sys.pathTok) {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "assertion "+clmAud+" does not contain "+sys.selfId+sys.pathTok, http.StatusBadRequest, nil))
	}

	// クライアント認証できた。
	log.Debug(sender, ": Authenticated "+req.ta())

	tokId := newId(sys.tokLen)

	// アクセストークンが決まった。
	log.Debug(sender, ": Generated token "+mosaic(tokId))

	// ID トークンの作成。
	acnt, err := sys.acntDb.Get(cod.Account())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		// アカウントが無い。
		return erro.Wrap(idperr.New(idperr.Invalid_request, "accout is not exist", http.StatusBadRequest, nil))
	}

	clms := map[string]interface{}{}
	if !cod.LoginDate().IsZero() {
		clms[clmAuth_time] = cod.LoginDate().Unix()
	}
	if cod.Nonce() != "" {
		clms[clmNonce] = cod.Nonce()
	}
	if hGen, err := jwt.HashFunction(sys.sigAlg); err != nil {
		return erro.Wrap(err)
	} else if hGen > 0 {
		h := hGen.New()
		h.Write([]byte(tokId))
		sum := h.Sum(nil)
		clms[clmAt_hash] = base64url.EncodeToString(sum[:len(sum)/2])
	}
	idTok, err := sys.newIdToken(ta, acnt, cod.IdTokenAttributes(), clms)
	if err != nil {
		return erro.Wrap(err)
	}

	// ID トークンができた。
	log.Debug(sender, ": Generated ID token")

	tok := token.New(
		tokId,
		now.Add(sys.tokExpIn),
		cod.Account(),
		cod.Scope(),
		cod.AccountAttributes(),
		cod.Ta(),
	)

	// アクセストークンを認可コードに結びつける。
	cod.SetToken(tokId)
	if ok, err := sys.acodDb.Replace(cod, savedCodDate); err != nil {
		return erro.Wrap(idperr.New(idperr.Server_error, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if !ok {
		disposeCode(sys, codId)
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+mosaic(codId)+" is used by others", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Linked token "+mosaic(tok.Id())+" to code "+mosaic(cod.Id()))

	// アクセストークンを保存する。
	if err := sys.tokDb.Save(tok, now.Add(sys.tokDbExpIn)); err != nil {
		return erro.Wrap(err)
	}

	log.Debug(sender, ": Saved token "+mosaic(tok.Id()))

	return responseToken(w, tok, "", idTok)
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
	cod, err := sys.acodDb.Get(codId)
	if err != nil {
		// 何もできない。
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		return
	} else if cod == nil {
		return
	}

	if cod.Token() == "" {
		return
	}
	disposeToken(sys, cod.Token())
}

// アクセストークンを廃棄処分する。
func disposeToken(sys *system, tokId string) {
	for {
		tok, err := sys.tokDb.Get(tokId)
		if err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
			return
		} else if tok == nil {
			return
		} else if tok.Invalid() {
			return
		}
		savedDate := tok.Date()

		tok.Invalidate()
		if ok, err := sys.tokDb.Replace(tok, savedDate); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
			return
		} else if ok {
			return
		}
	}
}
