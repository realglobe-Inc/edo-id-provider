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

// トークンエンドポイント。
package token

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	"github.com/realglobe-Inc/edo-id-provider/idputil"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"reflect"
	"time"
)

type handler struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	pathTok string

	pwSaltLen  int
	tokLen     int
	tokExpIn   time.Duration
	tokDbExpIn time.Duration
	jtiExpIn   time.Duration

	keyDb  keydb.Db
	acntDb account.Db
	taDb   tadb.Db
	sectDb sector.Db
	pwDb   pairwise.Db
	codDb  authcode.Db
	tokDb  token.Db
	jtiDb  jtidb.Db

	idGen rand.Generator
}

func New(
	stopper *server.Stopper,
	selfId string,
	sigAlg string,
	sigKid string,
	pathTok string,
	pwSaltLen int,
	tokLen int,
	tokExpIn time.Duration,
	tokDbExpIn time.Duration,
	jtiExpIn time.Duration,
	keyDb keydb.Db,
	acntDb account.Db,
	taDb tadb.Db,
	sectDb sector.Db,
	pwDb pairwise.Db,
	codDb authcode.Db,
	tokDb token.Db,
	jtiDb jtidb.Db,
	idGen rand.Generator,
) http.Handler {
	return &handler{
		stopper:    stopper,
		selfId:     selfId,
		sigAlg:     sigAlg,
		sigKid:     sigKid,
		pathTok:    pathTok,
		pwSaltLen:  pwSaltLen,
		tokLen:     tokLen,
		tokExpIn:   tokExpIn,
		tokDbExpIn: tokDbExpIn,
		jtiExpIn:   jtiExpIn,
		keyDb:      keyDb,
		acntDb:     acntDb,
		taDb:       taDb,
		sectDb:     sectDb,
		pwDb:       pwDb,
		codDb:      codDb,
		tokDb:      tokDb,
		jtiDb:      jtiDb,
		idGen:      idGen,
	}
}

func (this *handler) PairwiseSaltLength() int       { return this.pwSaltLen }
func (this *handler) SectorDb() sector.Db           { return this.sectDb }
func (this *handler) PairwiseDb() pairwise.Db       { return this.pwDb }
func (this *handler) IdGenerator() rand.Generator   { return this.idGen }
func (this *handler) KeyDb() keydb.Db               { return this.keyDb }
func (this *handler) SignAlgorithm() string         { return this.sigAlg }
func (this *handler) SignKeyId() string             { return this.sigKid }
func (this *handler) SelfId() string                { return this.selfId }
func (this *handler) JwtIdExpiresIn() time.Duration { return this.jtiExpIn }

// 主にテスト用。
func (this *handler) SetSelfId(selfId string) {
	this.selfId = selfId
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sender *requtil.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondApiError(w, r, erro.New(rcv), sender)
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, true)
	//////////////////////////////

	sender = requtil.Parse(r, "")
	log.Info(sender, ": Received token request")
	defer log.Info(sender, ": Handled token request")

	if err := this.serve(w, r, sender); err != nil {
		idperr.RespondApiError(w, r, erro.Wrap(err), sender)
		return
	}
}

func (this *handler) serve(w http.ResponseWriter, r *http.Request, sender *requtil.Request) error {
	if r.Method != tagPost {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported method "+r.Method, http.StatusMethodNotAllowed, nil))
	}

	req, err := parseRequest(r)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
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

	log.Debug(sender, ": Code "+logutil.Mosaic(codId)+" is declared")

	now := time.Now()

	cod, err := this.codDb.Get(codId)
	if err != nil {
		return erro.Wrap(err)
	} else if cod == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+logutil.Mosaic(codId)+" is not exist", http.StatusBadRequest, nil))
	} else if cod.Expires().Before(now) {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+logutil.Mosaic(codId)+" is expired", http.StatusBadRequest, nil))
	} else if cod.Token() != "" {
		disposeCode(this, codId)
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+logutil.Mosaic(codId)+" is invalid", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Code "+logutil.Mosaic(codId)+" is exist")
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

	if req.taAssertion() == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, "no "+tagClient_assertion, http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": "+tagClient_assertion+" is found")

	// Authorization ヘッダと client_secret パラメータも認識はする。
	if r.Header.Get(tagAuthorization) != "" || r.FormValue(tagClient_secret) != "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "multi client authentication algorithms are detected", http.StatusBadRequest, nil))
	}

	// クライアント認証する。
	ta, err := this.taDb.Get(req.ta())
	if err != nil {
		return erro.Wrap(err)
	} else if jti, err := idputil.VerifyAssertion(req.taAssertion(), ta.Id(), ta.Keys(), this.selfId+this.pathTok); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ok, err := this.jtiDb.SaveIfAbsent(jti); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.New("JWT ID overlaps")
	}

	// クライアント認証できた。
	log.Debug(sender, ": Authenticated "+req.ta())

	tokId := this.idGen.String(this.tokLen)

	// アクセストークンが決まった。
	log.Debug(sender, ": Generated token "+logutil.Mosaic(tokId))

	// ID トークンの作成。
	acnt, err := this.acntDb.Get(cod.Account())
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
	if hGen, err := jwt.HashFunction(this.sigAlg); err != nil {
		return erro.Wrap(err)
	} else {
		h := hGen.New()
		h.Write([]byte(tokId))
		sum := h.Sum(nil)
		clms[tagAt_hash] = base64url.EncodeToString(sum[:len(sum)/2])
	}
	idTok, err := idputil.IdToken(this, ta, acnt, cod.IdTokenAttributes(), clms)
	if err != nil {
		return erro.Wrap(err)
	}

	// ID トークンができた。
	log.Debug(sender, ": Generated ID token")

	tok := token.New(
		tokId,
		now.Add(this.tokExpIn),
		cod.Account(),
		cod.Scope(),
		cod.AccountAttributes(),
		cod.Ta(),
	)

	// アクセストークンを認可コードに結びつける。
	cod.SetToken(tokId)
	if ok, err := this.codDb.Replace(cod, savedCodDate); err != nil {
		return erro.Wrap(idperr.New(idperr.Server_error, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if !ok {
		disposeCode(this, codId)
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "code "+logutil.Mosaic(codId)+" is used by others", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Linked token "+logutil.Mosaic(tok.Id())+" to code "+logutil.Mosaic(cod.Id()))

	// アクセストークンを保存する。
	if err := this.tokDb.Save(tok, now.Add(this.tokDbExpIn)); err != nil {
		return erro.Wrap(err)
	}

	log.Debug(sender, ": Saved token "+logutil.Mosaic(tok.Id()))

	m := map[string]interface{}{
		tagAccess_token: tok.Id(),
		tagToken_type:   tagBearer,
	}
	m[tagExpires_in] = int64(this.tokExpIn / time.Second)
	if len(tok.Scope()) > 0 {
		m[tagScope] = requtil.ValueSetForm(tok.Scope())
	}
	if idTok != "" {
		m[tagId_token] = idTok
	}
	return idputil.RespondJson(w, m)
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
func disposeCode(this *handler, codId string) {
	cod, err := this.codDb.Get(codId)
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
	disposeToken(this, cod.Token())
}

// アクセストークンを廃棄処分する。
func disposeToken(this *handler, tokId string) {
	for {
		tok, err := this.tokDb.Get(tokId)
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
		if ok, err := this.tokDb.Replace(tok, savedDate); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
			return
		} else if ok {
			return
		}
	}
}
