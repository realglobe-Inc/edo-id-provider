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
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

func (sys *system) cooperateFromApi(w http.ResponseWriter, r *http.Request) error {
	sender := request.Parse(r, "")
	log.Info(sender, ": Received cooperation-from request")
	defer log.Info(sender, ": Handled cooperation-from request")

	req, err := parseCoopFromRequest(r)
	if err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	}

	switch req.grantType() {
	case tagAccess_token:
		return sys.cooperateFromServeAsMain(w, r, req, sender)
	case tagReferral:
		return sys.cooperateFromServeAsMain(w, r, req, sender)
	default:
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported grant "+req.grantType(), http.StatusBadRequest, nil))
	}
}

// 処理の主体が属す ID プロバイダとして対応。
func (sys *system) cooperateFromServeAsMain(w http.ResponseWriter, r *http.Request, req *coopFromRequest, sender *request.Request) error {
	if len(req.responseType()) > 2 || !req.responseType()[tagCode_token] {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported response type "+request.ValueSetForm(req.responseType()), http.StatusBadRequest, nil))
	}
	var reqRef bool
	if len(req.responseType()) == 1 {
		reqRef = false
	} else if reqRef = req.responseType()[tagReferral]; !reqRef {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported response type "+request.ValueSetForm(req.responseType()), http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Response types ", req.responseType(), " are OK")

	if req.fromTa() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no from-TA ID", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": From-TA "+req.fromTa()+" is declared")

	taFr, err := sys.taDb.Get(req.fromTa())
	if err != nil {
		return erro.Wrap(err)
	} else if taFr == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "from-TA "+req.fromTa()+" is not exist", http.StatusBadRequest, nil))
	} else if jti, err := sys.verifyTa(taFr, req.taAssertion(), sys.pathCoopFr); err != nil {
		return erro.Wrap(idperr.New(idperr.Invalid_client, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
	} else if ok, err := sys.jtiDb.SaveIfAbsent(jti); err != nil {
		return erro.Wrap(err)
	} else if !ok {
		return erro.New("JWT ID overlaps")
	}

	log.Debug(sender, ": Verified from-TA "+taFr.Id())

	if req.toTa() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no to-TA ID", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": To-TA "+req.fromTa()+" is declared")

	taTo, err := sys.taDb.Get(req.toTa())
	if err != nil {
		return erro.Wrap(err)
	} else if taTo == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "to-TA "+req.toTa()+" is not exist", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": To-TA "+taTo.Id()+" is exist")

	if req.grantType() != tagAccess_token {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported grant type "+req.grantType(), http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Grant type "+req.grantType()+" is OK")

	if req.accessToken() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no access token", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Access token "+mosaic(req.accessToken())+" is declared")

	now := time.Now()
	tok, err := sys.tokDb.Get(req.accessToken())
	if err != nil {
		return erro.Wrap(err)
	} else if tok == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "declared access token is not exist", http.StatusBadRequest, nil))
	} else if tok.Invalid() {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "access token is invalid", http.StatusBadRequest, nil))
	} else if now.After(tok.Expires()) {
		return erro.Wrap(idperr.New(idperr.Invalid_grant, "access token expired", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Access token "+mosaic(req.accessToken())+" is valid")

	var scop map[string]bool
	if req.scope() == nil {
		scop = tok.Scope()
		log.Debug(sender, ": Use token scope ", scop)
	} else {
		scop = req.scope()
		if !contains(tok.Scope(), scop) {
			return erro.Wrap(idperr.New(idperr.Invalid_scope, "not allowed scopes", http.StatusBadRequest, nil))
		}
		log.Debug(sender, ": Use given scope ", scop)
	}

	var exp time.Time
	if req.expiresIn() == 0 {
		exp = tok.Expires()
		log.Debug(sender, ": Use token expiration date")
	} else if exp = now.Add(req.expiresIn()); !tok.Expires().Before(exp) {
		log.Debug(sender, ": Use given token expiration duration")
	} else {
		log.Debug(sender, ": Use token expiration date")
		exp = tok.Expires()
	}

	if req.accountTag() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no main account tag", http.StatusBadRequest, nil))
	}

	log.Debug(sender, ": Main account tag is "+req.accountTag())

	allTags := map[string]bool{req.accountTag(): true}
	acntTags := map[string]bool{}
	if acnts := req.accounts(); len(acnts) > 0 {
		for tag, acntId := range acnts {
			if allTags[tag] {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "account tag "+tag+" overlaps", http.StatusBadRequest, nil))
			} else if acnt, err := sys.acntDb.Get(acntId); err != nil {
				return erro.Wrap(err)
			} else if acnt == nil {
				return erro.Wrap(idperr.New(idperr.Invalid_request, tag+" tagged account "+acntId+" is not exist", http.StatusBadRequest, nil))
			}

			allTags[tag] = true
			acntTags[tag] = true
			log.Debug(sender, ": "+tag+" tagged account "+acntId+" is exist")
		}
	}

	var keys []jwk.Key
	var ref []byte
	if reqRef {
		hashAlg := req.hashAlgorithm()
		if hashAlg == "" {
			hashAlg = sys.hashAlg
		}
		hashStrSize, err := hashStringSize(hashAlg)
		if err != nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "unsupported hash algorithm "+hashAlg, http.StatusBadRequest, nil))
		}

		log.Debug(sender, ": Hash algorithm "+hashAlg+" is OK")

		if len(req.relatedAccounts()) == 0 {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "no related accounts", http.StatusBadRequest, nil))
		}
		for tag, hashStr := range req.relatedAccounts() {
			if allTags[tag] {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "related account tag "+tag+" overlaps", http.StatusBadRequest, nil))
			} else if len(hashStr) != hashStrSize {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "invalid related account hash", http.StatusBadRequest, nil))
			}
			allTags[tag] = true
		}

		log.Debug(sender, ": Related accounts are OK")

		if len(req.relatedIdProviders()) == 0 {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "no related ID providers", http.StatusBadRequest, nil))
		}
		for _, idpId := range req.relatedIdProviders() {
			if idp, err := sys.idpDb.Get(idpId); err != nil {
				return erro.Wrap(err)
			} else if idp == nil {
				return erro.Wrap(idperr.New(idperr.Invalid_request, "no related ID provider "+idpId, http.StatusBadRequest, nil))
			}
		}

		jt := jwt.New()
		jt.SetHeader(tagAlg, sys.sigAlg)
		jt.SetClaim(tagIss, sys.selfId)
		jt.SetClaim(tagSub, taFr.Id())
		jt.SetClaim(tagAud, req.relatedIdProviders())
		jt.SetClaim(tagExp, now.Add(sys.jtiExpIn).Unix())
		jt.SetClaim(tagJti, randomString(sys.jtiLen))
		jt.SetClaim(tagTo_client, taTo.Id())
		jt.SetClaim(tagRelated_users, req.relatedAccounts())
		jt.SetClaim(tagHash_alg, hashAlg)

		if keys, err = sys.keyDb.Get(); err != nil {
			return erro.Wrap(err)
		} else if err := jt.Sign(keys); err != nil {
			return erro.Wrap(err)
		} else if ref, err = jt.Encode(); err != nil {
			return erro.Wrap(err)
		}

		log.Info(sender, ": Generated referral")
	}

	cod := randomString(sys.ccodLen)
	jt := jwt.New()
	jt.SetHeader(tagAlg, sys.sigAlg)
	jt.SetClaim(tagIss, sys.selfId)
	jt.SetClaim(tagSub, cod)
	jt.SetClaim(tagAud, taTo.Id())
	jt.SetClaim(tagUser_tag, req.accountTag())
	if len(acntTags) > 0 {
		jt.SetClaim(tagUser_tags, strset.Set(acntTags))
	}
	if ref != nil {
		hGen, err := jwt.HashFunction(sys.sigAlg)
		if err != nil {
			return erro.Wrap(err)
		}
		h := hGen.New()
		h.Write(ref)
		sum := h.Sum(nil)
		jt.SetClaim(tagRef_hash, base64url.EncodeToString(sum[:len(sum)/2]))
	}

	if keys == nil {
		if keys, err = sys.keyDb.Get(); err != nil {
			return erro.Wrap(err)
		}
	}
	if err := jt.Sign(keys); err != nil {
		return erro.Wrap(err)
	}
	codTok, err := jt.Encode()
	if err != nil {
		return erro.Wrap(err)
	}
	log.Info(sender, ": Generated code token")

	m := map[string]interface{}{
		tagCode_token: string(codTok),
	}
	if ref != nil {
		m[tagReferral] = string(ref)
	}
	return respondJson(w, m)
}

// 処理の主体が属さない ID プロバイダとして対応。
func (sys *system) cooperateFromServeAsSub(w http.ResponseWriter, r *http.Request, req *coopFromRequest, sender *request.Request) error {
	panic("not yet implemented")
}
