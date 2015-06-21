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

// TA 間連携連携先仲介エンドポイント。
package coopto

import (
	"github.com/realglobe-Inc/edo-id-provider/assertion"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	"github.com/realglobe-Inc/edo-id-provider/idputil"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"time"
)

type handler struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	pathCoopTo string

	pwSaltLen  int
	tokLen     int
	tokExpIn   time.Duration
	tokDbExpIn time.Duration
	jtiExpIn   time.Duration

	keyDb  keydb.Db
	acntDb account.Db
	consDb consent.Db
	taDb   tadb.Db
	sectDb sector.Db
	pwDb   pairwise.Db
	codDb  coopcode.Db
	tokDb  token.Db
	jtiDb  jtidb.Db
	idGen  rand.Generator

	debug bool
}

func New(
	stopper *server.Stopper,
	selfId string,
	sigAlg string,
	sigKid string,
	pathCoopTo string,
	pwSaltLen int,
	tokLen int,
	tokExpIn time.Duration,
	tokDbExpIn time.Duration,
	jtiExpIn time.Duration,
	keyDb keydb.Db,
	acntDb account.Db,
	consDb consent.Db,
	taDb tadb.Db,
	sectDb sector.Db,
	pwDb pairwise.Db,
	codDb coopcode.Db,
	tokDb token.Db,
	jtiDb jtidb.Db,
	idGen rand.Generator,
	debug bool,
) http.Handler {
	return &handler{
		stopper,
		selfId,
		sigAlg,
		sigKid,
		pathCoopTo,
		pwSaltLen,
		tokLen,
		tokExpIn,
		tokDbExpIn,
		jtiExpIn,
		keyDb,
		acntDb,
		consDb,
		taDb,
		sectDb,
		pwDb,
		codDb,
		tokDb,
		jtiDb,
		idGen,
		debug,
	}
}

func (this *handler) PairwiseSaltLength() int     { return this.pwSaltLen }
func (this *handler) SectorDb() sector.Db         { return this.sectDb }
func (this *handler) PairwiseDb() pairwise.Db     { return this.pwDb }
func (this *handler) IdGenerator() rand.Generator { return this.idGen }

// 主にテスト用。
func (this *handler) SetSelfId(selfId string) {
	this.selfId = selfId
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var sender *requtil.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondJson(w, r, erro.New(rcv), sender)
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, this.debug)
	//////////////////////////////

	sender = requtil.Parse(r, "")
	log.Info(sender, ": Received cooperation-to request")
	defer log.Info(sender, ": Handled cooperation-to request")

	if err := (&environment{this, sender}).serve(w, r); err != nil {
		idperr.RespondJson(w, r, erro.Wrap(err), sender)
		return
	}
}

// environment のメソッドは idperr.Error を返す。
type environment struct {
	*handler

	sender *requtil.Request
}

func (this *environment) serve(w http.ResponseWriter, r *http.Request) error {
	req, err := parseRequest(r)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if req.grantType() != tagCooperation_code {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "invalid grant type "+req.grantType(), http.StatusBadRequest, nil))
	}

	cod, err := this.codDb.Get(req.code())
	if err != nil {
		return erro.Wrap(err)
	} else if cod == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "declared code is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(this.sender, ": Declared code is exist")

	toTa, err := this.taDb.Get(cod.ToTa())
	if err != nil {
		return erro.Wrap(err)
	} else if toTa == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "to-TA "+cod.ToTa()+" is not exist", http.StatusBadRequest, nil))
	} else if ass, err := assertion.Parse(req.taAssertion()); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ass.Issuer() != toTa.Id() {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, ass.Issuer()+" is not code holder", http.StatusBadRequest, nil))
	} else if err := ass.Verify(toTa.Id(), toTa.Keys(), this.selfId+this.pathCoopTo); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ok, err := this.jtiDb.SaveIfAbsent(jtidb.New(toTa.Id(), ass.Id(), ass.Expires())); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.New("JWT ID overlaps")
	}

	log.Debug(this.sender, ": Verified to-TA "+toTa.Id())

	tagToId := map[string]string{}
	for _, acnt := range cod.Accounts() {
		tagToId[acnt.Tag()] = acnt.Id()
	}
	for tag := range req.subClaims() {
		if tagToId[tag] == "" {
			return erro.Wrap(idperr.New(idperr.Invalid_grant, "invalid user tag "+tag, http.StatusBadRequest, nil))
		}
	}

	if cod.SourceToken() != "" {
		return this.serveAsMain(w, r, req, cod, toTa)
	} else {
		return this.serveAsSub(w, r, req, cod, toTa)
	}
}

func (this *environment) serveAsMain(w http.ResponseWriter, r *http.Request, req *request, cod *coopcode.Element, toTa tadb.Element) error {
	return this.serveAs(true, w, r, req, cod, toTa)
}

func (this *environment) serveAsSub(w http.ResponseWriter, r *http.Request, req *request, cod *coopcode.Element, toTa tadb.Element) error {
	return this.serveAs(false, w, r, req, cod, toTa)
}

func (this *environment) serveAs(isMain bool, w http.ResponseWriter, r *http.Request, req *request, cod *coopcode.Element, toTa tadb.Element) error {
	ids := map[string]map[string]interface{}{}

	var tok *token.Element
	if isMain {
		acnt, err := this.acntDb.Get(cod.Account().Id())
		if err != nil {
			return erro.Wrap(err)
		} else if acnt == nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "accout "+cod.Account().Tag()+" is not exist", http.StatusForbidden, nil))
		} else if err := idputil.SetSub(this, acnt, toTa); err != nil {
			return erro.Wrap(err)
		} else if err := idputil.CheckClaims(acnt, req.claims().IdTokenEntries()); err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		} else if err := idputil.CheckClaims(acnt, req.claims().AccountEntries()); err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		}
		cons, err := this.consDb.Get(acnt.Id(), toTa.Id())
		if err != nil {
			return erro.Wrap(err)
		} else if cons == nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, "accout "+cod.Account().Tag()+" allowed nothing", http.StatusForbidden, nil))
		}

		scop, err := idputil.ProvidedScopes(cons.Scope(), cod.Scope())
		if err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		}

		acntAttrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), scop, req.claims().AccountEntries())
		if err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		}
		tok = token.New(
			this.idGen.String(this.tokLen),
			cod.TokenExpires(),
			cod.Account().Id(),
			cod.Scope(),
			acntAttrs,
			toTa.Id(),
		)
		idTokAttrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), nil, req.claims().IdTokenEntries())
		if err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		}
		m := map[string]interface{}{}
		for attr := range idTokAttrs {
			m[attr] = acnt.Attribute(attr)
		}
		ids[cod.Account().Tag()] = m

		log.Debug(this.sender, ": Account "+cod.Account().Tag()+" allowed required attributes")
	}

	for _, codAcnt := range cod.Accounts() {
		acnt, err := this.acntDb.Get(codAcnt.Id())
		if err != nil {
			return erro.Wrap(err)
		} else if acnt == nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "accout "+codAcnt.Tag()+" is not exist", http.StatusForbidden, nil))
		} else if err := idputil.SetSub(this, acnt, toTa); err != nil {
			return erro.Wrap(err)
		} else if err := idputil.CheckClaims(acnt, req.subClaims()[codAcnt.Tag()]); err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		}
		cons, err := this.consDb.Get(acnt.Id(), toTa.Id())
		if err != nil {
			return erro.Wrap(err)
		} else if cons == nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, "account "+codAcnt.Tag()+" allowed nothing", http.StatusForbidden, nil))
		}

		attrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), nil, req.subClaims()[codAcnt.Tag()])
		if err != nil {
			return erro.Wrap(idperr.New(idperr.Access_denied, erro.Unwrap(err).Error(), http.StatusForbidden, err))
		}
		m := map[string]interface{}{}
		for attr := range attrs {
			m[attr] = acnt.Attribute(attr)
		}
		ids[codAcnt.Tag()] = m

		log.Debug(this.sender, ": Account "+codAcnt.Tag()+" allowed required attributes")
	}

	now := time.Now()
	jt := jwt.New()
	jt.SetHeader(tagAlg, this.sigAlg)
	if this.sigKid != "" {
		jt.SetHeader(tagKid, this.sigKid)
	}
	jt.SetClaim(tagIss, this.selfId)
	jt.SetClaim(tagSub, cod.FromTa())
	jt.SetClaim(tagAud, audience.New(toTa.Id()))
	jt.SetClaim(tagExp, now.Add(this.jtiExpIn).Unix())
	jt.SetClaim(tagIat, now.Unix())
	jt.SetClaim(tagIds, ids)
	if keys, err := this.keyDb.Get(); err != nil {
		return erro.Wrap(err)
	} else if err := jt.Sign(keys); err != nil {
		return erro.Wrap(err)
	}
	idsTok, err := jt.Encode()
	if err != nil {
		return erro.Wrap(err)
	}

	respParams := map[string]interface{}{
		tagIds_token: string(idsTok),
	}

	if isMain {
		respParams[tagAccess_token] = tok.Id()
		respParams[tagToken_type] = tagBearer

		// TODO アクセストークンを元になったアクセストークンに紐付ける。
		log.Warn(this.sender, ": Not linked token "+logutil.Mosaic(tok.Id())+" to source token "+logutil.Mosaic(cod.SourceToken()))

		// アクセストークンを保存する。
		if err := this.tokDb.Save(tok, tok.Expires().Add(this.tokDbExpIn-this.tokExpIn)); err != nil {
			return erro.Wrap(err)
		}

		log.Debug(this.sender, ": Saved token "+logutil.Mosaic(tok.Id()))

		if !tok.Expires().IsZero() {
			respParams[tagExpires_in] = int64(tok.Expires().Sub(time.Now()).Seconds())
		}
		if len(tok.Scope()) > 0 {
			respParams[tagScope] = requtil.ValueSetForm(tok.Scope())
		}
	}
	return idputil.RespondJson(w, respParams)
}
