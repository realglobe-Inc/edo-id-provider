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
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"time"
)

// ログイン UI にリダイレクトする。
func (sys *system) redirectToLoginUi(w http.ResponseWriter, r *http.Request, sess *session.Element, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(sys.pathLginUi)
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sess)
	}

	// ログインページに渡すクエリパラメータを生成。
	q := uri.Query()
	q.Set(formIssuer, sys.selfId)
	if acnts := sess.SelectedAccounts(); len(acnts) > 0 {
		q.Set(formUsernames, accountSetForm(acnts))
	}
	if disp := sess.Request().Display(); disp != "" {
		q.Set(formDisplay, disp)
	}
	if lang, langs := sess.Language(), sess.Request().Languages(); lang != "" || len(langs) > 0 {
		q.Set(formLocales, languagesForm(lang, langs))
	}
	if msg != "" {
		q.Set(formMessage, msg)
	}
	uri.RawQuery = q.Encode()

	// チケットを発行。
	uri.Fragment = newId(sys.ticLen)
	sess.SetTicket(uri.Fragment)
	log.Info("Ticket " + mosaic(uri.Fragment) + " was published")

	log.Info("Redirect " + mosaic(sess.Id()) + " to login UI")
	return sys.redirectTo(w, r, uri, sess)
}

// ログイン UI からの入力を受け付けて続きをする。
func (sys *system) lginPage(w http.ResponseWriter, r *http.Request) (err error) {

	var sess *session.Element
	if sessId := newBaseRequest(r).session(); sessId != "" {
		// セッションが通知された。
		log.Debug("Session " + mosaic(sessId) + " is declared")

		if sess, err = sys.sessDb.Get(sessId); err != nil {
			log.Err(erro.Wrap(err))
			// 新規発行すれば動くので諦めない。
		} else if sess == nil {
			// セッションが無かった。
			log.Warn("Declared session " + mosaic(sessId) + " is not exist")
		} else {
			// セッションがあった。
			log.Debug("Declared session " + mosaic(sessId) + " is exist")
		}
	}

	now := time.Now()
	if sess == nil || now.After(sess.Expires()) {
		sess = session.New(newId(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info("New session " + mosaic(sess.Id()) + " was generated but not yet registered")
	}

	// セッションは決まった。

	authReq := sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return sys.returnError(w, r, idperr.New(idperr.Invalid_request, "session "+mosaic(sess.Id())+" is not in authentication process", http.StatusBadRequest, nil), sess)
	}

	// ユーザー認証中。
	log.Debug("session " + mosaic(sess.Id()) + " is in authentication process")

	req := newLoginRequest(r)
	if sess.Ticket() == "" {
		// ログイン中でない。
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Access_denied, "not in interactive process", nil), sess)
	} else if req.ticket() != sess.Ticket() {
		// 無効なログイン券。
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Access_denied, "invalid ticket "+mosaic(req.ticket()), nil), sess)
	}

	// チケットが有効だった。
	log.Debug("Ticket " + mosaic(req.ticket()) + " is OK")

	if req.accountName() == "" || req.passInfo() == nil {
		// ログイン情報不備。
		log.Debug("Login info is not specified")
		return sys.redirectToLoginUi(w, r, sess, "Please log in")
	}

	// ログイン情報があった。
	log.Debug("Login info is specified")

	acnt, err := sys.acntDb.GetByName(req.accountName())
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sess)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug("Specified accout " + req.accountName() + " was not found")
		return sys.redirectToLoginUi(w, r, sess, "Accout "+req.accountName()+" was not found. Please log in")
	} else if req.passType() != acnt.Authenticator().Type() {
		return sys.redirectToLoginUi(w, r, sess, "Not registered password type. Please log in")
	} else if pass := req.passInfo(); pass == nil {
		return sys.redirectToLoginUi(w, r, sess, "Some required info is lost. Please log in")
	} else if !acnt.Authenticator().Verify(pass.password(), pass.params()...) {
		// パスワード間違い。
		return sys.redirectToLoginUi(w, r, sess, "Wrong password. Please log in")
	}

	// ログインできた。
	log.Info("Account " + acnt.Id() + " (" + acnt.Name() + ") logged in")

	sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	sess.Account().Login()
	if lang := req.language(); lang != "" {
		sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug("Language " + lang + " was selected")
	}

	return sys.afterLogin(w, r, sess, nil, acnt)
}

// ログインが終わったところから。
func (sys *system) afterLogin(w http.ResponseWriter, r *http.Request, sess *session.Element, ta tadb.Element, acnt account.Element) (err error) {

	if ta == nil {
		if ta, err = sys.taDb.Get(sess.Request().Ta()); err != nil {
			return sys.redirectError(w, r, erro.Wrap(err), sess)
		} else if ta == nil {
			// アカウントが無い。
			return sys.redirectError(w, r, newErrorForRedirect(idperr.Server_error, "TA "+mosaic(sess.Request().Ta())+" was not found", nil), sess)
		}
	}
	if acnt == nil {
		if acnt, err = sys.acntDb.Get(sess.Account().Id()); err != nil {
			return sys.redirectError(w, r, erro.Wrap(err), sess)
		} else if acnt == nil {
			// アカウントが無い。
			return sys.redirectError(w, r, newErrorForRedirect(idperr.Server_error, "accout "+mosaic(sess.Account().Id())+" was not found", nil), sess)
		}
	}

	// クレーム指定の検査。
	if err := sys.setSub(acnt, ta); err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sess)
	} else if err := checkContradiction(acnt, sess.Request().Claims()); err != nil {
		// 指定を満たすのは無理。
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), erro.Wrap(err)), sess)
	}

	prmpts := sess.Request().Prompt()
	if prmpts[prmptConsent] {
		if prmpts[prmptNone] {
			return sys.redirectError(w, r, newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil), sess)
		}

		return sys.redirectToConsentUi(w, r, sess, "Please allow to provide these scope and attributes")
	}

	// 事前同意を調べる。
	cons, err := sys.consDb.Get(sess.Account().Id(), sess.Request().Ta())
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sess)
	} else if cons == nil {
		cons = consent.New(sess.Account().Id(), sess.Request().Ta())
	}

	if ok, scop, tokAttrs, acntAttrs := satisfiable(cons, removeUnknownScope(sess.Request().Scope()), sess.Request().Claims()); ok {
		// 事前同意で十分。
		log.Debug("Already consented")
		return sys.afterConsent(w, r, sess, ta, acnt, scop, tokAttrs, acntAttrs)
	}

	if prmpts[prmptNone] {
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil), sess)
	}

	return sys.redirectToConsentUi(w, r, sess, "Please allow to provide these scope and attributes")
}
