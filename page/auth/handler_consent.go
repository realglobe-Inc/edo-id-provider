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

package auth

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	hashutil "github.com/realglobe-Inc/edo-id-provider/hash"
	"github.com/realglobe-Inc/edo-id-provider/idputil"
	"github.com/realglobe-Inc/edo-id-provider/scope"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-idp-selector/ticket"
	"github.com/realglobe-Inc/edo-lib/jwt"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
)

// 同意 UI からの入力を受け付けて続きをする。
func (this *Page) HandleConsent(w http.ResponseWriter, r *http.Request) {
	var logPref string

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondHtml(w, r, erro.New(rcv), this.errTmpl, logPref)
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	sender := request.Parse(r, this.sessLabel)
	logPref = sender.String() + ": "

	server.LogRequest(level.DEBUG, r, this.debug, logPref)

	log.Info(logPref, "Received consent request")
	defer log.Info(logPref, "Handled consent request")

	var sess *session.Element
	if sessId := sender.Session(); sessId != "" {
		// セッションが通知された。
		log.Debug(logPref, "Session is declared")

		var err error
		if sess, err = this.sessDb.Get(sessId); err != nil {
			log.Err(logPref, erro.Wrap(err))
			// 新規発行すれば動くので諦めない。
		} else if sess == nil {
			// セッションが無かった。
			log.Warn(logPref, "Declared session is not exist")
		} else {
			// セッションがあった。
			log.Debug(logPref, "Declared session is exist")
		}
	}

	if now := time.Now(); sess == nil || now.After(sess.Expires()) {
		sess = session.New(this.idGen.String(this.sessLen), now.Add(this.sessExpIn))
		log.Info(logPref, "Generated new session "+logutil.Mosaic(sess.Id())+" but not yet registered")
	}

	// セッションが決まった。

	env := (&environment{this, logPref, sess})
	if err := env.consentServe(w, r); err != nil {
		env.respondErrorHtml(w, r, erro.Wrap(err))
		return
	}
}

// 同意 UI にリダイレクトする。
func (this *environment) redirectToConsentUi(w http.ResponseWriter, r *http.Request, msg string) error {

	uri, err := url.Parse(this.pathConsUi)
	if err != nil {
		return erro.Wrap(err)
	}

	// 同意ページに渡すクエリパラメータを生成。
	q := uri.Query()
	q.Set(tagIssuer, this.selfId)
	q.Set(tagUsername, this.sess.Account().Name())
	if scop := scope.RemoveUnknown(this.sess.Request().Scope()); len(scop) > 0 {
		q.Set(tagScope, request.ValueSetForm(scop))
	}
	if this.sess.Request().Claims() != nil {
		attrs, optAttrs := this.sess.Request().Claims().Names()
		if len(attrs) > 0 {
			q.Set(tagClaims, request.ValueSetForm(attrs))
		}
		if len(optAttrs) > 0 {
			q.Set(tagOptional_claims, request.ValueSetForm(optAttrs))
		}
	}
	q.Set(tagExpires_in, strconv.FormatInt(int64(this.tokExpIn/time.Second), 10))
	q.Set(tagClient_id, this.sess.Request().Ta())
	if disp := this.sess.Request().Display(); disp != "" {
		q.Set(tagDisplay, disp)
	}
	if lang, langs := this.sess.Language(), this.sess.Request().Languages(); lang != "" || len(langs) > 0 {
		q.Set(tagLocales, languagesForm(lang, langs))
	}
	if msg != "" {
		q.Set(tagMessage, msg)
	}
	uri.RawQuery = q.Encode()

	// チケットを発行。
	uri.Fragment = this.idGen.String(this.ticLen)
	this.sess.SetTicket(ticket.New(uri.Fragment, time.Now().Add(this.ticExpIn)))
	log.Info(this.logPref, "Published ticket "+logutil.Mosaic(uri.Fragment))

	log.Info(this.logPref, "Redirect to consent UI")
	this.redirectTo(w, r, uri)
	return nil
}

func (this *environment) consentServe(w http.ResponseWriter, r *http.Request) error {
	authReq := this.sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil))
	}

	// ユーザー認証・認可処理中。
	log.Debug(this.logPref, "Session is in authentication process")

	req, err := parseConsentRequest(r)
	if err != nil {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), err))
	} else if tic := this.sess.Ticket(); tic == nil {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "no consent session", nil))
	} else if req.ticket() != tic.Id() {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "invalid ticket", nil))
	} else if tic.Expires().Before(time.Now()) {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "ticket expired", nil))
	}

	// チケットが有効だった。
	log.Debug(this.logPref, "Ticket "+logutil.Mosaic(req.ticket())+" is OK")

	scopCons := consent.Consent(req.allowedScope())
	scop, err := idputil.ProvidedScopes(scopCons, scope.RemoveUnknown(this.sess.Request().Scope()))
	if err != nil {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), err))
	}
	attrCons := consent.Consent(req.allowedAttributes())
	idTokAttrs, err := idputil.ProvidedAttributes(scopCons, attrCons, nil, this.sess.Request().Claims().IdTokenEntries())
	if err != nil {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), err))
	}
	acntAttrs, err := idputil.ProvidedAttributes(scopCons, attrCons, scop, this.sess.Request().Claims().AccountEntries())
	if err != nil {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), err))
	}

	// 同意できた。
	log.Debug(this.logPref, "Essential claims were allowed")

	// 同意情報を更新。
	cons, err := this.consDb.Get(this.sess.Account().Id(), this.sess.Request().Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if cons == nil {
		cons = consent.New(this.sess.Account().Id(), this.sess.Request().Ta())
	}

	for v := range req.allowedScope() {
		cons.Scope().SetAllow(v)
	}
	for v := range req.allowedAttributes() {
		cons.Attribute().SetAllow(v)
	}
	for v := range req.deniedScope() {
		cons.Scope().SetDeny(v)
	}
	for v := range req.deniedAttributes() {
		cons.Attribute().SetDeny(v)
	}
	if err := this.consDb.Save(cons); err != nil {
		log.Err(this.logPref, erro.Wrap(err))
		// 今回だけは動くので諦めない。
	}

	if loc := req.language(); loc != "" {
		this.sess.SetLanguage(loc)

		// 言語を選択してた。
		log.Debug(this.logPref, "Locale "+loc+" was selected")
	}

	return this.afterConsent(w, r, nil, nil, scop, idTokAttrs, acntAttrs)
}

// 同意が終わったところから。
func (this *environment) afterConsent(w http.ResponseWriter, r *http.Request, ta tadb.Element, acnt account.Element, scop, idTokAttrs, acntAttrs map[string]bool) (err error) {

	if !scop[tagOpenid] {
		// openid すら許されなかった。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, tagOpenid+" scope was denied", nil))
	}

	req := this.sess.Request()

	// 認可コードを発行する。
	now := time.Now()
	cod := authcode.New(this.idGen.String(this.codLen), now.Add(this.codExpIn), this.sess.Account().Id(),
		this.sess.Account().LoginDate(), scop, idTokAttrs, acntAttrs, req.Ta(),
		req.RedirectUri(), req.Nonce())

	log.Debug(this.logPref, "Generated authorization code "+logutil.Mosaic(cod.Id()))

	if err := this.codDb.Save(cod, now.Add(this.codDbExpIn)); err != nil {
		return erro.Wrap(err)
	}

	// 認可コードを発行した。
	log.Info(this.logPref, "Published authorization code "+logutil.Mosaic(cod.Id())+" to "+logutil.Mosaic(this.sess.Id()))

	var idTok string
	if req.ResponseType()[tagId_token] {
		if ta == nil {
			if ta, err = this.taDb.Get(this.sess.Request().Ta()); err != nil {
				return erro.Wrap(err)
			} else if ta == nil {
				// TA が無い。
				return erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "TA "+this.sess.Request().Ta()+" is not exist", nil))
			}
		}
		if acnt == nil {
			if acnt, err = this.acntDb.Get(this.sess.Account().Id()); err != nil {
				return erro.Wrap(err)
			} else if acnt == nil {
				// アカウントが無い。
				return erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "accout is not exist", nil))
			}
		}

		clms := map[string]interface{}{}
		if this.sess.Request().MaxAge() >= 0 {
			clms[tagAuth_time] = this.sess.Account().LoginDate().Unix()
		}
		if cod.Nonce() != "" {
			clms[tagNonce] = cod.Nonce()
		}
		if hGen := jwt.HashGenerator(this.sigAlg); !hGen.Available() {
			return erro.New("unsupported algorithm " + this.sigAlg)
		} else if hGen > 0 {
			clms[tagC_hash] = hashutil.Hashing(hGen.New(), []byte(cod.Id()))
		}
		idTok, err = idputil.IdToken(this, ta, acnt, idTokAttrs, clms)
		if err != nil {
			return erro.Wrap(err)
		}
	}

	this.redirectCode(w, r, cod, idTok)
	return nil
}
