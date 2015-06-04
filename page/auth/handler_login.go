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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/edo-id-provider/idputil"
	"github.com/realglobe-Inc/edo-id-provider/scope"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/url"
	"time"
)

// ログイン UI からの入力を受け付けて続きをする。
func (this *Page) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var sender *request.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondPageError(w, r, erro.New(rcv), sender, this.errTmpl)
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

	sender = request.Parse(r, this.sessLabel)
	log.Info(sender, ": Received login request")
	defer log.Info(sender, ": Handled login request")

	if err := this.loginServe(w, r, sender); err != nil {
		idperr.RespondPageError(w, r, erro.Wrap(err), sender, this.errTmpl)
		return
	}
	return
}

func (this *Page) loginServe(w http.ResponseWriter, r *http.Request, sender *request.Request) (err error) {
	var sess *session.Element
	if sessId := sender.Session(); sessId != "" {
		// セッションが通知された。
		log.Debug(sender, ": Session is declared")

		if sess, err = this.sessDb.Get(sessId); err != nil {
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

	if now := time.Now(); sess == nil || now.After(sess.Expires()) {
		sess = session.New(this.idGen.String(this.sessLen), now.Add(this.sessExpIn))
		log.Info(sender, ": Generated new session "+logutil.Mosaic(sess.Id())+" but not yet registered")
	}

	// セッションは決まった。

	if err := this.loginServeWithSession(w, r, sender, sess); err != nil {
		return this.respondPageError(w, r, erro.Wrap(err), sender, sess)
	}
	return nil
}

// ログイン UI にリダイレクトする。
func (this *Page) redirectToLoginUi(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(this.pathLginUi)
	if err != nil {
		return this.respondPageError(w, r, erro.Wrap(err), sender, sess)
	}

	// ログインページに渡すクエリパラメータを生成。
	q := uri.Query()
	q.Set(tagIssuer, this.selfId)
	if acnts := sess.SelectedAccounts(); len(acnts) > 0 {
		q.Set(tagUsernames, accountSetForm(acnts))
	}
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
	uri.Fragment = this.idGen.String(this.ticLen)
	sess.SetTicket(uri.Fragment)
	log.Info(sender, ": Published ticket "+logutil.Mosaic(uri.Fragment))

	log.Info(sender, ": Redirect to login UI")
	return this.redirectTo(w, r, uri, sender, sess)
}

func (this *Page) loginServeWithSession(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element) error {
	authReq := sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil))
	}

	// ユーザー認証中。
	log.Debug(sender, ": Session is in authentication process")

	req := newLoginRequest(r)
	if sess.Ticket() == "" {
		// ログイン中でない。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "not in interactive process", nil))
	} else if req.ticket() != sess.Ticket() {
		// 無効なログイン券。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "invalid ticket "+logutil.Mosaic(req.ticket()), nil))
	}

	// チケットが有効だった。
	log.Debug(sender, ": Ticket "+logutil.Mosaic(req.ticket())+" is OK")

	if req.accountName() == "" || req.passInfo() == nil {
		// ログイン情報不備。
		log.Debug(sender, ": No login info")
		return this.redirectToLoginUi(w, r, sender, sess, "Please log in")
	}

	// ログイン情報があった。
	log.Debug(sender, ": Login info is specified")

	acnt, err := this.acntDb.GetByName(req.accountName())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug(sender, ": Specified accout "+req.accountName()+" is not exist")
		return this.redirectToLoginUi(w, r, sender, sess, "Accout "+req.accountName()+" is not exist. Please log in")
	} else if req.passType() != acnt.Authenticator().Type() {
		return this.redirectToLoginUi(w, r, sender, sess, "Not registered password type. Please log in")
	} else if pass := req.passInfo(); pass == nil {
		return this.redirectToLoginUi(w, r, sender, sess, "No required info. Please log in")
	} else if !acnt.Authenticator().Verify(pass.password(), pass.params()...) {
		// パスワード間違い。
		return this.redirectToLoginUi(w, r, sender, sess, "Wrong password. Please log in")
	}

	// ログインできた。
	log.Info(sender, ": Account "+acnt.Id()+" ("+acnt.Name()+") logged in")

	sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	sess.Account().Login()
	if lang := req.language(); lang != "" {
		sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug(sender, ": Language "+lang+" was selected")
	}

	return this.afterLogin(w, r, sender, sess, nil, acnt)
}

// ログインが終わったところから。
func (this *Page) afterLogin(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, ta tadb.Element, acnt account.Element) (err error) {

	if ta == nil {
		if ta, err = this.taDb.Get(sess.Request().Ta()); err != nil {
			return erro.Wrap(err)
		} else if ta == nil {
			// TA が無い。
			return erro.Wrap(newErrorForRedirect(idperr.Server_error, "TA "+logutil.Mosaic(sess.Request().Ta())+" is not exist", nil))
		}
	}
	if acnt == nil {
		if acnt, err = this.acntDb.Get(sess.Account().Id()); err != nil {
			return erro.Wrap(err)
		} else if acnt == nil {
			// アカウントが無い。
			return erro.Wrap(newErrorForRedirect(idperr.Server_error, "accout is not exist", nil))
		}
	}

	// クレーム指定の検査。
	if err := idputil.SetSub(this, acnt, ta); err != nil {
		return erro.Wrap(err)
	} else if err := idputil.CheckClaims(acnt, sess.Request().Claims().IdTokenEntries()); err != nil {
		// 指定を満たすのは無理。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), erro.Wrap(err)))
	} else if err := idputil.CheckClaims(acnt, sess.Request().Claims().AccountEntries()); err != nil {
		// 指定を満たすのは無理。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), erro.Wrap(err)))
	}

	prmpts := sess.Request().Prompt()
	if prmpts[tagConsent] {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(sender, ": Consent is forced")
		return this.redirectToConsentUi(w, r, sender, sess, "Please allow to provide these scope and attributes")
	}

	// 事前同意を調べる。
	cons, err := this.consDb.Get(sess.Account().Id(), sess.Request().Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if cons == nil {
		cons = consent.New(sess.Account().Id(), sess.Request().Ta())
	}

	scop, err := idputil.ProvidedScopes(cons.Scope(), scope.RemoveUnknown(sess.Request().Scope()))
	if err != nil {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(sender, ": Consent is required: ", erro.Unwrap(err))
		return this.redirectToConsentUi(w, r, sender, sess, "Please allow to provide these scope and attributes")
	}
	idTokAttrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), nil, sess.Request().Claims().IdTokenEntries())
	if err != nil {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(sender, ": Consent is required: ", erro.Unwrap(err))
		return this.redirectToConsentUi(w, r, sender, sess, "Please allow to provide these scope and attributes")
	}
	acntAttrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), scop, sess.Request().Claims().AccountEntries())
	if err != nil {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(sender, ": Consent is required: ", erro.Unwrap(err))
		return this.redirectToConsentUi(w, r, sender, sess, "Please allow to provide these scope and attributes")
	}

	// 事前同意で十分。
	log.Debug(sender, ": Already consented")
	return this.afterConsent(w, r, sender, sess, ta, acnt, scop, idTokAttrs, acntAttrs)
}
