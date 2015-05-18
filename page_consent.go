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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/edo-id-provider/request"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-lib/base64url"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// 同意 UI にリダイレクトする。
func (sys *system) redirectToConsentUi(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, msg string) error {

	uri, err := url.Parse(sys.pathConsUi)
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
	}

	// 同意ページに渡すクエリパラメータを生成。
	q := uri.Query()
	q.Set(tagIssuer, sys.selfId)
	q.Set(tagUsername, sess.Account().Name())
	if scop := removeUnknownScope(sess.Request().Scope()); len(scop) > 0 {
		q.Set(tagScope, request.ValueSetForm(scop))
	}
	if sess.Request().Claims() != nil {
		attrs, optAttrs := sess.Request().Claims().Names()
		if len(attrs) > 0 {
			q.Set(tagClaims, request.ValueSetForm(attrs))
		}
		if len(optAttrs) > 0 {
			q.Set(tagOptional_claims, request.ValueSetForm(optAttrs))
		}
	}
	q.Set(tagExpires_in, strconv.FormatInt(int64(sys.tokExpIn/time.Second), 10))
	q.Set(tagClient_id, sess.Request().Ta())
	if disp := sess.Request().Display(); disp != "" {
		q.Set(tagDisplay, disp)
	}
	if lang, langs := sess.Language(), sess.Request().Languages(); lang != "" || len(langs) > 0 {
		q.Set(tagLocales, languagesForm(lang, langs))
	}
	if msg != "" {
		q.Set(tagMessage, msg)
	}
	uri.RawQuery = q.Encode()

	// チケットを発行。
	uri.Fragment = newId(sys.ticLen)
	sess.SetTicket(uri.Fragment)
	log.Info(sender, ": Published ticket "+mosaic(uri.Fragment))

	log.Info(sender, ": Redirect to consent UI")
	return sys.redirectTo(w, r, uri, sender, sess)
}

// 同意 UI からの入力を受け付けて続きをする。
func (sys *system) consentPage(w http.ResponseWriter, r *http.Request) (err error) {
	sender := request.Parse(r, sys.sessLabel)

	var sess *session.Element
	if sessId := sender.Session(); sessId != "" {
		// セッションが通知された。
		log.Debug(sender, ": Session is declared")

		if sess, err = sys.sessDb.Get(sessId); err != nil {
			log.Err(sender, ": ", erro.Wrap(err))
			// 新規発行すれば動くので諦めない。
		} else if sess == nil {
			// セッションが無かった。
			log.Warn(sender, ": Declared session is not exist")
		} else {
			// セッションがあった。
			log.Debug(sender, ": Declared session is exist")
		}
	}

	now := time.Now()
	if sess == nil || now.After(sess.Expires()) {
		sess = session.New(newId(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info(sender, ": Generated new session "+mosaic(sess.Id())+" but not yet registered")
	}

	// セッションは決まった。

	authReq := sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return sys.returnError(w, r, erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil)), sender, sess)
	}

	// ユーザー認証・認可処理中。
	log.Debug(sender, ": Session is in authentication process")

	req := newConsentRequest(r)
	if sess.Ticket() == "" {
		// 同意中でない。
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Access_denied, "not in interaction process", nil)), sender, sess)
	} else if req.ticket() != sess.Ticket() {
		// 無効な同意券。
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Access_denied, "invalid ticket "+mosaic(req.ticket()), nil)), sender, sess)
	}

	// チケットが有効だった。
	log.Debug(sender, ": Ticket "+mosaic(req.ticket())+" is OK")

	ok, scop, tokAttrs, acntAttrs := satisfiable(newConsentInfo(req.allowedScope(), req.allowedAttributes()), removeUnknownScope(sess.Request().Scope()), sess.Request().Claims())
	if !ok {
		// 同意が足りなかった。
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Access_denied, "some essential claims were denied", nil)), sender, sess)
	}

	// 同意できた。
	log.Debug(sender, ": Essential claims were allowed")

	// 同意情報を更新。
	cons, err := sys.consDb.Get(sess.Account().Id(), sess.Request().Ta())
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
	} else if cons == nil {
		cons = consent.New(sess.Account().Id(), sess.Request().Ta())
	}

	for v := range req.allowedScope() {
		cons.AllowScope(v)
	}
	for v := range req.allowedAttributes() {
		cons.AllowAttribute(v)
	}
	for v := range req.deniedScope() {
		cons.DenyScope(v)
	}
	for v := range req.deniedAttributes() {
		cons.DenyAttribute(v)
	}
	if err := sys.consDb.Save(cons); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
		// 今回だけは動くので諦めない。
	}

	if loc := req.language(); loc != "" {
		sess.SetLanguage(loc)

		// 言語を選択してた。
		log.Debug(sender, ": Locale "+loc+" was selected")
	}

	return sys.afterConsent(w, r, sender, sess, nil, nil, scop, tokAttrs, acntAttrs)
}

// 同意が終わったところから。
func (sys *system) afterConsent(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, ta tadb.Element, acnt account.Element, scop, tokAttrs, acntAttrs map[string]bool) (err error) {

	if !scop[tagOpenid] {
		// openid すら許されなかった。
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Access_denied, tagOpenid+" scope was denied", nil)), sender, sess)
	}

	req := sess.Request()

	// 認可コードを発行する。
	now := time.Now()
	cod := authcode.New(newId(sys.acodLen), now.Add(sys.acodExpIn), sess.Account().Id(),
		sess.Account().LoginDate(), scop, tokAttrs, acntAttrs, req.Ta(),
		req.RedirectUri(), req.Nonce())

	log.Debug(sender, ": Generated authorization code "+mosaic(cod.Id()))

	if err := sys.acodDb.Save(cod, now.Add(sys.acodDbExpIn)); err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
	}

	// 認可コードを発行した。
	log.Info(sender, ": Published authorization code "+mosaic(cod.Id())+" to "+mosaic(sess.Id()))

	var idTok string
	if req.ResponseType()[tagId_token] {
		if ta == nil {
			if ta, err = sys.taDb.Get(sess.Request().Ta()); err != nil {
				return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
			} else if ta == nil {
				// TA が無い。
				return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "TA "+sess.Request().Ta()+" is not exist", nil)), sender, sess)
			}
		}
		if acnt == nil {
			if acnt, err = sys.acntDb.Get(sess.Account().Id()); err != nil {
				return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
			} else if acnt == nil {
				// アカウントが無い。
				return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "accout is not exist", nil)), sender, sess)
			}
		}

		clms := map[string]interface{}{}
		if sess.Request().MaxAge() >= 0 {
			clms[tagAuth_time] = sess.Account().LoginDate().Unix()
		}
		if cod.Nonce() != "" {
			clms[tagNonce] = cod.Nonce()
		}
		if hGen, err := jwt.HashFunction(sys.sigAlg); err != nil {
			return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
		} else if hGen > 0 {
			h := hGen.New()
			h.Write([]byte(cod.Id()))
			sum := h.Sum(nil)
			clms[tagC_hash] = base64url.EncodeToString(sum[:len(sum)/2])
		}
		idTok, err = sys.newIdToken(ta, acnt, tokAttrs, clms)
		if err != nil {
			return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
		}
	}

	return sys.redirectCode(w, r, cod, idTok, sender, sess)
}
