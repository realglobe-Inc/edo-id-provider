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
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"time"
)

// ログイン UI にリダイレクトする。
func (sys *system) redirectToLoginUi(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(sys.pathLginUi)
	if err != nil {
		return sys.respondPageError(w, r, erro.Wrap(err), sender, sess)
	}

	// ログインページに渡すクエリパラメータを生成。
	q := uri.Query()
	q.Set(tagIssuer, sys.selfId)
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
	uri.Fragment = randomString(sys.ticLen)
	sess.SetTicket(uri.Fragment)
	log.Info(sender, ": Published ticket "+mosaic(uri.Fragment))

	log.Info(sender, ": Redirect to login UI")
	return sys.redirectTo(w, r, uri, sender, sess)
}

// ログイン UI からの入力を受け付けて続きをする。
func (sys *system) lginPage(w http.ResponseWriter, r *http.Request) (err error) {
	sender := request.Parse(r, sys.sessLabel)
	log.Info(sender, ": Received login request")
	defer log.Info(sender, ": Handled login request")

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

	if now := time.Now(); sess == nil || now.After(sess.Expires()) {
		sess = session.New(randomString(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info(sender, ": Generated new session "+mosaic(sess.Id())+" but not yet registered")
	}

	// セッションは決まった。

	if err := sys.loginServe(w, r, sender, sess); err != nil {
		return sys.respondPageError(w, r, erro.Wrap(err), sender, sess)
	}
	return nil
}

func (sys *system) loginServe(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element) error {
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
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "invalid ticket "+mosaic(req.ticket()), nil))
	}

	// チケットが有効だった。
	log.Debug(sender, ": Ticket "+mosaic(req.ticket())+" is OK")

	if req.accountName() == "" || req.passInfo() == nil {
		// ログイン情報不備。
		log.Debug(sender, ": No login info")
		return sys.redirectToLoginUi(w, r, sender, sess, "Please log in")
	}

	// ログイン情報があった。
	log.Debug(sender, ": Login info is specified")

	acnt, err := sys.acntDb.GetByName(req.accountName())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug(sender, ": Specified accout "+req.accountName()+" is not exist")
		return sys.redirectToLoginUi(w, r, sender, sess, "Accout "+req.accountName()+" is not exist. Please log in")
	} else if req.passType() != acnt.Authenticator().Type() {
		return sys.redirectToLoginUi(w, r, sender, sess, "Not registered password type. Please log in")
	} else if pass := req.passInfo(); pass == nil {
		return sys.redirectToLoginUi(w, r, sender, sess, "No required info. Please log in")
	} else if !acnt.Authenticator().Verify(pass.password(), pass.params()...) {
		// パスワード間違い。
		return sys.redirectToLoginUi(w, r, sender, sess, "Wrong password. Please log in")
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

	return sys.afterLogin(w, r, sender, sess, nil, acnt)
}

// ログインが終わったところから。
func (sys *system) afterLogin(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, ta tadb.Element, acnt account.Element) (err error) {

	if ta == nil {
		if ta, err = sys.taDb.Get(sess.Request().Ta()); err != nil {
			return erro.Wrap(err)
		} else if ta == nil {
			// アカウントが無い。
			return erro.Wrap(newErrorForRedirect(idperr.Server_error, "TA "+mosaic(sess.Request().Ta())+" is not exist", nil))
		}
	}
	if acnt == nil {
		if acnt, err = sys.acntDb.Get(sess.Account().Id()); err != nil {
			return erro.Wrap(err)
		} else if acnt == nil {
			// アカウントが無い。
			return erro.Wrap(newErrorForRedirect(idperr.Server_error, "accout is not exist", nil))
		}
	}

	// クレーム指定の検査。
	if err := sys.setSub(acnt, ta); err != nil {
		return erro.Wrap(err)
	} else if err := checkContradiction(acnt, sess.Request().Claims()); err != nil {
		// 指定を満たすのは無理。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), erro.Wrap(err)))
	}

	prmpts := sess.Request().Prompt()
	if prmpts[tagConsent] {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
		}

		return sys.redirectToConsentUi(w, r, sender, sess, "Please allow to provide these scope and attributes")
	}

	// 事前同意を調べる。
	cons, err := sys.consDb.Get(sess.Account().Id(), sess.Request().Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if cons == nil {
		cons = consent.New(sess.Account().Id(), sess.Request().Ta())
	}

	if ok, scop, tokAttrs, acntAttrs := satisfiable(cons, removeUnknownScope(sess.Request().Scope()), sess.Request().Claims()); ok {
		// 事前同意で十分。
		log.Debug(sender, ": Already consented")
		return sys.afterConsent(w, r, sender, sess, ta, acnt, scop, tokAttrs, acntAttrs)
	}

	if prmpts[tagNone] {
		return erro.Wrap(newErrorForRedirect(idperr.Consent_required, "cannot consent without UI", nil))
	}

	return sys.redirectToConsentUi(w, r, sender, sess, "Please allow to provide these scope and attributes")
}
