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

// TA 間連携連携元仲介エンドポイント。
package coopfrom

import (
	"hash"
	"net/http"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/assertion"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	hashutil "github.com/realglobe-Inc/edo-id-provider/hash"
	"github.com/realglobe-Inc/edo-id-provider/idputil"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
)

type handler struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	pathCoopFr string

	codLen     int
	codExpIn   time.Duration
	codDbExpIn time.Duration
	jtiLen     int
	jtiExpIn   time.Duration

	keyDb  keydb.Db
	pwDb   pairwise.Db
	acntDb account.Db
	taDb   tadb.Db
	idpDb  idpdb.Db
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
	pathCoopFr string,
	codLen int,
	codExpIn time.Duration,
	codDbExpIn time.Duration,
	jtiLen int,
	jtiExpIn time.Duration,
	keyDb keydb.Db,
	pwDb pairwise.Db,
	acntDb account.Db,
	taDb tadb.Db,
	idpDb idpdb.Db,
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
		pathCoopFr,
		codLen,
		codExpIn,
		codDbExpIn,
		jtiLen,
		jtiExpIn,
		keyDb,
		pwDb,
		acntDb,
		taDb,
		idpDb,
		codDb,
		tokDb,
		jtiDb,
		idGen,
		debug,
	}
}

// 主にテスト用。
func (this *handler) SetSelfId(selfId string) {
	this.selfId = selfId
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var logPref string

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondJson(w, r, erro.New(rcv), logPref)
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	logPref = server.ParseSender(r) + ": "

	server.LogRequest(level.DEBUG, r, this.debug, logPref)

	log.Info(logPref, "Received cooperation-from request")
	defer log.Info(logPref, "Handled cooperation-from request")

	if err := (&environment{this, logPref}).serve(w, r); err != nil {
		idperr.RespondJson(w, r, erro.Wrap(err), logPref)
		return
	}
}

// environment のメソッドは idperr.Error を返す。
type environment struct {
	*handler

	logPref string
}

func (this *environment) serve(w http.ResponseWriter, r *http.Request) error {
	req, err := parseRequest(r)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	}

	switch req.grantType() {
	case tagAccess_token:
		return this.serveAsMain(w, r, req)
	case tagReferral:
		return this.serveAsSub(w, r, req)
	default:
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported grant "+req.grantType(), http.StatusBadRequest, nil))
	}
}

// 処理の主体が属す ID プロバイダとして対応。
func (this *environment) serveAsMain(w http.ResponseWriter, r *http.Request, req *request) error {
	if len(req.responseType()) > 2 || !req.responseType()[tagCode_token] {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported response type "+requtil.ValueSetForm(req.responseType()), http.StatusBadRequest, nil))
	}
	var reqRef bool
	if len(req.responseType()) == 1 {
		reqRef = false
	} else if reqRef = req.responseType()[tagReferral]; !reqRef {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported response type "+requtil.ValueSetForm(req.responseType()), http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Response types "+requtil.ValueSetForm(req.responseType())+" are OK")

	if req.fromTa() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no from-TA ID", http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "From-TA "+req.fromTa()+" is declared")

	frTa, err := this.taDb.Get(req.fromTa())
	if err != nil {
		return erro.Wrap(err)
	} else if frTa == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "from-TA "+req.fromTa()+" is not exist", http.StatusBadRequest, nil))
	} else if ass, err := assertion.Parse(req.taAssertion()); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if err := ass.Verify(frTa.Id(), frTa.Keys(), this.selfId+this.pathCoopFr); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ok, err := this.jtiDb.SaveIfAbsent(jtidb.New(frTa.Id(), ass.Id(), ass.Expires())); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.New("JWT ID overlaps")
	}

	log.Debug(this.logPref, "Verified from-TA "+frTa.Id())

	if req.toTa() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no to-TA ID", http.StatusBadRequest, nil))
	} else if req.toTa() == req.fromTa() {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "to-TA is from-TA", http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "To-TA "+req.fromTa()+" is declared")

	toTa, err := this.taDb.Get(req.toTa())
	if err != nil {
		return erro.Wrap(err)
	} else if toTa == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "to-TA "+req.toTa()+" is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "To-TA "+toTa.Id()+" is exist")

	if req.grantType() != tagAccess_token {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported grant type "+req.grantType(), http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Grant type "+req.grantType()+" is OK")

	if req.accessToken() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no access token", http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Access token "+logutil.Mosaic(req.accessToken())+" is declared")

	now := time.Now()
	tok, err := this.tokDb.Get(req.accessToken())
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "declared access token is not exist", http.StatusBadRequest, nil))
	} else if tok.Invalid() {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "access token is invalid", http.StatusBadRequest, nil))
	} else if now.After(tok.Expires()) {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "access token expired", http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Access token "+logutil.Mosaic(req.accessToken())+" is valid")

	var scop map[string]bool
	if req.scope() == nil {
		scop = tok.Scope()
		log.Debug(this.logPref, "Use token scope ", scop)
	} else {
		scop = req.scope()
		if !strsetutil.Contains(tok.Scope(), scop) {
			return erro.Wrap(idperr.New(idperr.Invalid_scope, "not allowed scopes", http.StatusBadRequest, nil))
		}
		log.Debug(this.logPref, "Use given scope ", scop)
	}

	var exp time.Time
	if req.expiresIn() == 0 {
		exp = tok.Expires()
		log.Debug(this.logPref, "Use token expiration date")
	} else if exp = now.Add(req.expiresIn()); !tok.Expires().Before(exp) {
		log.Debug(this.logPref, "Use given token expiration duration")
	} else {
		log.Debug(this.logPref, "Use token expiration date")
		exp = tok.Expires()
	}

	if req.accountTag() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no main account tag", http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Main account tag is "+req.accountTag())

	codAcnts, err := this.getAccounts(req.accounts(), frTa)
	if err != nil {
		return erro.Wrap(err)
	}
	allTags := map[string]bool{req.accountTag(): true}
	for _, codAcnt := range codAcnts {
		if allTags[codAcnt.Tag()] {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "account tag "+codAcnt.Tag()+" overlaps", http.StatusBadRequest, nil))
		}
		allTags[codAcnt.Tag()] = true
	}

	var keys []jwk.Key
	var ref []byte
	var hFun hash.Hash
	if reqRef {
		for acntTag := range req.relatedAccounts() {
			if allTags[acntTag] {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "related account tag "+acntTag+" overlaps", http.StatusBadRequest, nil))
			}
			allTags[acntTag] = true
		}
		if keys, err = this.keyDb.Get(); err != nil {
			return erro.Wrap(err)
		} else if ref, err = this.makeReferral(req, keys); err != nil {
			return erro.Wrap(err)
		}
		log.Info(this.logPref, "Generated referral")

		hGen := hashutil.Generator(req.hashAlgorithm())
		if !hGen.Available() {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported hash algorithm "+req.hashAlgorithm(), http.StatusBadRequest, nil))
		}
		hFun = hGen.New()
	}

	codId := this.idGen.String(this.codLen)
	if keys == nil {
		if keys, err = this.keyDb.Get(); err != nil {
			return erro.Wrap(err)
		}
	}
	codTok, err := makeCodeToken(this.handler, req, codId, frTa.Id(), toTa.Id(), ref, hFun, keys)
	if err != nil {
		return erro.Wrap(err)
	}
	log.Info(this.logPref, "Generated code token")

	cod := coopcode.New(codId, now.Add(this.codExpIn), coopcode.NewAccount(tok.Account(), req.accountTag()), tok.Id(), scop, exp, codAcnts, frTa.Id(), toTa.Id())
	if err := this.codDb.Save(cod, now.Add(this.codDbExpIn)); err != nil {
		return erro.Wrap(err)
	}
	log.Info(this.logPref, "Saved code")

	m := map[string]interface{}{
		tagCode_token: string(codTok),
	}
	if ref != nil {
		m[tagReferral] = string(ref)
	}
	return idputil.RespondJson(w, m)
}

// 処理の主体が属さない ID プロバイダとして対応。
func (this *environment) serveAsSub(w http.ResponseWriter, r *http.Request, req *request) error {
	if len(req.responseType()) > 1 || !req.responseType()[tagCode_token] {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported response type "+requtil.ValueSetForm(req.responseType()), http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Response types "+requtil.ValueSetForm(req.responseType())+" is OK")

	if req.grantType() != tagReferral {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported grant type "+req.grantType(), http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Grant type "+req.grantType()+" is OK")

	if req.referral() == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no referral", http.StatusBadRequest, nil))
	} else if len(req.accounts()) == 0 {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no accounts", http.StatusBadRequest, nil))
	}

	ref, err := parseReferral(req.referral())
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	}

	log.Debug(this.logPref, "Parsed referral")

	idp, err := this.idpDb.Get(ref.idProvider())
	if err != nil {
		return erro.Wrap(err)
	} else if idp == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if err := ref.verify(idp.Keys(), this.selfId); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	}

	log.Debug(this.logPref, "Primary ID provider "+idp.Id()+" is exist")

	frTa, err := this.taDb.Get(ref.fromTa())
	if err != nil {
		return erro.Wrap(err)
	} else if frTa == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "from-TA "+ref.fromTa()+" is not exist", http.StatusBadRequest, err))
	}

	log.Debug(this.logPref, "From-TA "+frTa.Id()+" is exist")

	if ass, err := assertion.Parse(req.taAssertion()); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if err := ass.Verify(frTa.Id(), frTa.Keys(), this.selfId+this.pathCoopFr); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ok, err := this.jtiDb.SaveIfAbsent(jtidb.New(frTa.Id(), ass.Id(), ass.Expires())); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.New("JWT ID overlaps")
	}

	log.Debug(this.logPref, "Authenticated from-TA "+frTa.Id())

	toTa, err := this.taDb.Get(ref.toTa())
	if err != nil {
		return erro.Wrap(err)
	} else if toTa == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "to-TA "+ref.toTa()+" is not exist", http.StatusBadRequest, err))
	}

	log.Debug(this.logPref, "To-TA "+frTa.Id()+" is exist")

	codAcnts, err := this.getAccounts(req.accounts(), frTa)
	if err != nil {
		return erro.Wrap(err)
	}

	log.Debug(this.logPref, "Accounts are exist")

	hGen := hashutil.Generator(ref.hashAlgorithm())
	if !hGen.Available() {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "unsupported hash algorithm "+ref.hashAlgorithm(), http.StatusBadRequest, nil))
	}
	hFun := hGen.New()
	for acntTag, acntId := range req.accounts() {
		hFun.Reset()
		if hashutil.Hashing(hFun, []byte(this.selfId), []byte{0}, []byte(acntId)) != ref.relatedAccounts()[acntTag] {
			return erro.Wrap(idperr.New(idperr.Invalid_grant, "invalid account hash", http.StatusBadRequest, nil))
		}
	}

	log.Debug(this.logPref, "Account hashes are OK")

	codId := this.idGen.String(this.codLen)
	keys, err := this.keyDb.Get()
	if err != nil {
		return erro.Wrap(err)
	}
	hFun.Reset()
	codTok, err := makeCodeToken(this.handler, req, codId, frTa.Id(), toTa.Id(), req.referral(), hFun, keys)
	if err != nil {
		return erro.Wrap(err)
	}
	log.Info(this.logPref, "Generated code token")

	now := time.Now()
	cod := coopcode.New(codId, now.Add(this.codExpIn), nil, "", nil, time.Time{}, codAcnts, frTa.Id(), toTa.Id())
	if err := this.codDb.Save(cod, now.Add(this.codDbExpIn)); err != nil {
		return erro.Wrap(err)
	}
	log.Info(this.logPref, "Saved code")

	return idputil.RespondJson(w, map[string]interface{}{
		tagCode_token: string(codTok),
	})
}

// リクエストの users パラメータに対応するアカウント情報を返す。
// 返り値はアカウントタグからアカウント情報へのマップ。
func (this *environment) getAccounts(tagToId map[string]string, frTa tadb.Element) ([]*coopcode.Account, error) {
	codAcnts := []*coopcode.Account{}
	for acntTag, acntId := range tagToId {
		if frTa.Pairwise() {
			pw, err := this.pwDb.GetByPairwise(frTa.Sector(), acntId)
			if err != nil {
				return nil, erro.Wrap(err)
			} else if pw == nil {
				return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "no pairwise ID", http.StatusBadRequest, nil))
			}
			acntId = pw.Account()
		}
		acnt, err := this.acntDb.Get(acntId)
		if err != nil {
			return nil, erro.Wrap(err)
		} else if acnt == nil {
			return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "account "+acntId+" tagged by "+acntTag+" is not exist", http.StatusBadRequest, nil))
		}

		codAcnts = append(codAcnts, coopcode.NewAccount(acntId, acntTag))
	}
	return codAcnts, nil
}

func (this *environment) makeReferral(req *request, keys []jwk.Key) ([]byte, error) {
	hashStrSize := hashutil.Size(req.hashAlgorithm())
	if hashStrSize == 0 {
		return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported hash algorithm "+req.hashAlgorithm(), http.StatusBadRequest, nil))
	}

	log.Debug(this.logPref, "Hash algorithm "+req.hashAlgorithm()+" is supported")

	if len(req.relatedAccounts()) == 0 {
		return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "no related accounts", http.StatusBadRequest, nil))
	}
	for _, acntHash := range req.relatedAccounts() {
		if len(acntHash) != hashStrSize {
			return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "invalid related account hash", http.StatusBadRequest, nil))
		}
	}

	log.Debug(this.logPref, "Related accounts are OK")

	if len(req.relatedIdProviders()) == 0 {
		return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "no related ID providers", http.StatusBadRequest, nil))
	}
	for _, idpId := range req.relatedIdProviders() {
		if idp, err := this.idpDb.Get(idpId); err != nil {
			return nil, erro.Wrap(err)
		} else if idp == nil {
			return nil, erro.Wrap(idperr.New(idperr.Invalid_request, "no related ID provider "+idpId, http.StatusBadRequest, nil))
		}
	}

	ref := jwt.New()
	ref.SetHeader(tagAlg, this.sigAlg)
	if this.sigKid != "" {
		ref.SetHeader(tagKid, this.sigKid)
	}
	ref.SetClaim(tagIss, this.selfId)
	ref.SetClaim(tagSub, req.fromTa())
	ref.SetClaim(tagAud, req.relatedIdProviders())
	ref.SetClaim(tagExp, time.Now().Add(this.jtiExpIn).Unix())
	ref.SetClaim(tagJti, this.idGen.String(this.jtiLen))
	ref.SetClaim(tagTo_client, req.toTa())
	ref.SetClaim(tagRelated_users, req.relatedAccounts())
	ref.SetClaim(tagHash_alg, req.hashAlgorithm())

	if err := ref.Sign(keys); err != nil {
		return nil, erro.Wrap(err)
	}
	data, err := ref.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}

	return data, nil
}

func makeCodeToken(hndl *handler, req *request, codId, frTa, toTa string, ref []byte, hFun hash.Hash, keys []jwk.Key) ([]byte, error) {
	jt := jwt.New()
	jt.SetHeader(tagAlg, hndl.sigAlg)
	if hndl.sigKid != "" {
		jt.SetHeader(tagKid, hndl.sigKid)
	}
	jt.SetClaim(tagIss, hndl.selfId)
	jt.SetClaim(tagSub, codId)
	jt.SetClaim(tagAud, toTa)
	jt.SetClaim(tagFrom_client, frTa)
	if req.accountTag() != "" {
		jt.SetClaim(tagUser_tag, req.accountTag())
	}
	if len(req.accounts()) > 0 {
		acntTags := []string{}
		for acntTag := range req.accounts() {
			acntTags = append(acntTags, acntTag)
		}
		jt.SetClaim(tagUser_tags, acntTags)
	}
	if ref != nil {
		jt.SetClaim(tagRef_hash, hashutil.Hashing(hFun, ref))
	}

	if err := jt.Sign(keys); err != nil {
		return nil, erro.Wrap(err)
	}
	data, err := jt.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}

	return data, nil
}
