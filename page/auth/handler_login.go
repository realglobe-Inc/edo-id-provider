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
	"github.com/realglobe-Inc/edo-idp-selector/ticket"
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

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, this.debug)
	//////////////////////////////

	sender := request.Parse(r, this.sessLabel)
	logPref = sender.String() + ": "

	log.Info(logPref, "Received login request")
	defer log.Info(logPref, "Handled login request")

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

	// セッションは決まった。

	env := &environment{this, logPref, sess}
	if err := env.loginServe(w, r); err != nil {
		env.respondErrorHtml(w, r, erro.Wrap(err))
		return
	}
}

// ログイン UI にリダイレクトする。
func (this *environment) redirectToLoginUi(w http.ResponseWriter, r *http.Request, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(this.pathLginUi)
	if err != nil {
		return erro.Wrap(err)
	}

	// ログインページに渡すクエリパラメータを生成。
	q := uri.Query()
	q.Set(tagIssuer, this.selfId)
	if acnts := this.sess.SelectedAccounts(); len(acnts) > 0 {
		q.Set(tagUsernames, accountSetForm(acnts))
	}
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

	log.Info(this.logPref, "Redirect to login UI")
	this.redirectTo(w, r, uri)
	return nil
}

func (this *environment) loginServe(w http.ResponseWriter, r *http.Request) error {
	authReq := this.sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil))
	}

	// ユーザー認証中。
	log.Debug(this.logPref, "Session is in authentication process")

	req, err := parseLoginRequest(r)
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

	if req.accountName() == "" || req.passInfo() == nil {
		// ログイン情報不備。
		log.Debug(this.logPref, "No login info")
		return this.redirectToLoginUi(w, r, "Please log in")
	}

	// ログイン情報があった。
	log.Debug(this.logPref, "Login info is specified")

	acnt, err := this.acntDb.GetByName(req.accountName())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug(this.logPref, "Specified accout "+req.accountName()+" is not exist")
		return this.redirectToLoginUi(w, r, "Accout "+req.accountName()+" is not exist. Please log in")
	} else if err := acnt.Authenticator().Verify(req.passInfo().params()...); err != nil {
		// パスワード間違い。
		log.Warn(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		return this.redirectToLoginUi(w, r, "Wrong password. Please log in")
	}

	// ログインできた。
	log.Info(this.logPref, "Account "+acnt.Id()+" ("+acnt.Name()+") logged in")

	this.sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	this.sess.Account().Login()
	if lang := req.language(); lang != "" {
		this.sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug(this.logPref, "Language "+lang+" was selected")
	}

	return this.afterLogin(w, r, nil, acnt)
}

// ログインが終わったところから。
func (this *environment) afterLogin(w http.ResponseWriter, r *http.Request, ta tadb.Element, acnt account.Element) (err error) {

	if ta == nil {
		if ta, err = this.taDb.Get(this.sess.Request().Ta()); err != nil {
			return erro.Wrap(err)
		} else if ta == nil {
			// TA が無い。
			return erro.Wrap(newErrorForRedirect(idperr.Server_error, "TA "+logutil.Mosaic(this.sess.Request().Ta())+" is not exist", nil))
		}
	}
	if acnt == nil {
		if acnt, err = this.acntDb.Get(this.sess.Account().Id()); err != nil {
			return erro.Wrap(err)
		} else if acnt == nil {
			// アカウントが無い。
			return erro.Wrap(newErrorForRedirect(idperr.Server_error, "accout is not exist", nil))
		}
	}

	// クレーム指定の検査。
	if err := idputil.SetSub(this, acnt, ta); err != nil {
		return erro.Wrap(err)
	} else if err := idputil.CheckClaims(acnt, this.sess.Request().Claims().IdTokenEntries()); err != nil {
		// 指定を満たすのは無理。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), erro.Wrap(err)))
	} else if err := idputil.CheckClaims(acnt, this.sess.Request().Claims().AccountEntries()); err != nil {
		// 指定を満たすのは無理。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), erro.Wrap(err)))
	}

	prmpts := this.sess.Request().Prompt()
	if prmpts[tagConsent] {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(this.logPref, "Consent is forced")
		return this.redirectToConsentUi(w, r, "Please allow to provide these scope and attributes")
	}

	// 事前同意を調べる。
	cons, err := this.consDb.Get(this.sess.Account().Id(), this.sess.Request().Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if cons == nil {
		cons = consent.New(this.sess.Account().Id(), this.sess.Request().Ta())
	}

	scop, err := idputil.ProvidedScopes(cons.Scope(), scope.RemoveUnknown(this.sess.Request().Scope()))
	if err != nil {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(this.logPref, "Consent is required: ", erro.Unwrap(err))
		return this.redirectToConsentUi(w, r, "Please allow to provide these scope and attributes")
	}
	idTokAttrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), nil, this.sess.Request().Claims().IdTokenEntries())
	if err != nil {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(this.logPref, "Consent is required: ", erro.Unwrap(err))
		return this.redirectToConsentUi(w, r, "Please allow to provide these scope and attributes")
	}
	acntAttrs, err := idputil.ProvidedAttributes(cons.Scope(), cons.Attribute(), scop, this.sess.Request().Claims().AccountEntries())
	if err != nil {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		log.Debug(this.logPref, "Consent is required: ", erro.Unwrap(err))
		return this.redirectToConsentUi(w, r, "Please allow to provide these scope and attributes")
	}

	// 事前同意で十分。
	log.Debug(this.logPref, "Already consented")
	return this.afterConsent(w, r, ta, acnt, scop, idTokAttrs, acntAttrs)
}
