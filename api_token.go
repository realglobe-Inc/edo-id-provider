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
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"reflect"
	"time"
)

func respondToken(w http.ResponseWriter, tok *token.Element, refTok, idTok string) error {
	m := map[string]interface{}{
		tagAccess_token: tok.Id(),
		tagToken_type:   tagBearer,
	}
	if !tok.Expires().IsZero() {
		m[tagExpires_in] = int64(tok.Expires().Sub(time.Now()).Seconds())
	}
	if len(tok.Scope()) > 0 {
		var buff string
		for scop := range tok.Scope() {
			if len(buff) > 0 {
				buff += " "
			}
			buff += scop
		}
		m[tagScope] = buff
	}
	if refTok != "" {
		m[tagRefresh_token] = refTok
	}
	if idTok != "" {
		m[tagId_token] = idTok
	}
	return respondJson(w, m)
}

func (sys *system) tokenApi(w http.ResponseWriter, r *http.Request) error {
	sender := request.Parse(r, "")
	log.Info(sender, ": Received token request")
	defer log.Info(sender, ": Handled token request")

	if r.Method != tagPost {
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
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+tagGrant_type, http.StatusBadRequest, nil))
	} else if grntType != tagAuthorization_code {
		return erro.Wrap(idperr.New(idperr.Unsupported_grant_type, "unsupported grant type "+grntType, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Grant type is "+tagAuthorization_code)

	codId := req.code()
	if codId == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+tagCode, http.StatusBadRequest, nil))
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
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+tagClient_id, http.StatusBadRequest, nil))
	} else if req.ta() != cod.Ta() {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "you are not code holder", http.StatusBadRequest, nil))
	} else {
		log.Debug(sender, ": TA "+req.ta()+" is declared")
	}

	rediUri := req.redirectUri()
	if rediUri == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no "+tagRedirect_uri, http.StatusBadRequest, nil))
	} else if !reflect.DeepEqual(rediUri, cod.RedirectUri()) {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "invalid "+tagRedirect_uri, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+tagRedirect_uri+" matches that of code")

	if taAssType := req.taAssertionType(); taAssType == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no "+tagClient_assertion_type, http.StatusBadRequest, nil))
	} else if taAssType != cliAssTypeJwt_bearer {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "unsupported assertion type "+taAssType, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+tagClient_assertion_type+" is "+cliAssTypeJwt_bearer)

	taAss := req.taAssertion()
	if taAss == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no "+tagClient_assertion, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+tagClient_assertion+" is found")

	// Authorization ヘッダと client_secret パラメータも認識はする。
	if r.Header.Get(tagAuthorization) != "" || r.FormValue(tagClient_secret) != "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "multi client authentication algorithms are detected", http.StatusBadRequest, nil))
	}

	// クライアント認証する。
	ta, err := sys.taDb.Get(req.ta())
	if err != nil {
		return erro.Wrap(err)
	} else if jti, err := sys.verifyTa(ta, req.taAssertion()); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ok, err := sys.jtiDb.SaveIfAbsent(jti); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.New("JWT ID overlaps")
	}

	// クライアント認証できた。
	log.Debug(sender, ": Authenticated "+req.ta())

	tokId := randomString(sys.tokLen)

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
		clms[tagAuth_time] = cod.LoginDate().Unix()
	}
	if cod.Nonce() != "" {
		clms[tagNonce] = cod.Nonce()
	}
	if hGen, err := jwt.HashFunction(sys.sigAlg); err != nil {
		return erro.Wrap(err)
	} else if hGen > 0 {
		h := hGen.New()
		h.Write([]byte(tokId))
		sum := h.Sum(nil)
		clms[tagAt_hash] = base64url.EncodeToString(sum[:len(sum)/2])
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

	return respondToken(w, tok, "", idTok)
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
