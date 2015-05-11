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
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	jsonutil "github.com/realglobe-Inc/edo-lib/json"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"time"
)

// アカウント名の JSON 配列にする。
func accountSetForm(acnts []*session.Account) string {
	acntNames := []string{}
	for _, acnt := range acnts {
		acntNames = append(acntNames, acnt.Name())
	}

	buff, err := json.Marshal(acntNames)
	if err == nil {
		return string(buff)
	}
	log.Err(erro.Wrap(err))

	// 最後の手段。
	buff = []byte{'['}
	for _, v := range acntNames {
		if len(buff) > 2 {
			buff = append(buff, ',')
		}
		buff = append(buff, '"')
		buff = append(buff, jsonutil.Escape([]byte(v))...)
		buff = append(buff, '"')
	}
	buff = append(buff, ']')
	return string(buff)
}

// 空白区切りの文字列にする。
func languagesForm(lang string, langs []string) string {
	a := []string{}
	m := map[string]bool{}
	for _, v := range append([]string{lang}, langs...) {
		if v == "" || m[v] {
			continue
		}
		a = append(a, v)
		m[v] = true
	}
	return valuesForm(a)
}

// アカウント選択 UI にリダイレクトする。
func (sys *system) redirectToSelectUi(w http.ResponseWriter, r *http.Request, sess *session.Element, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(sys.pathSelUi)
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sess)
	}

	// アカウント選択ページに渡すクエリパラメータを生成。
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

	log.Info("Redirect " + mosaic(sess.Id()) + " to select UI")
	return sys.redirectTo(w, r, uri, sess)
}

// アカウント UI ページからの入力を受け付けて続きをする。
func (sys *system) selectPage(w http.ResponseWriter, r *http.Request) (err error) {

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
		// セッションを新規発行。
		sess = session.New(newId(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info("New session " + mosaic(sess.Id()) + " was generated but not yet saved")
	}

	// セッションは決まった。

	authReq := sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return sys.returnError(w, r, idperr.New(idperr.Invalid_request, "session "+mosaic(sess.Id())+" is not in authentication process", http.StatusBadRequest, nil), sess)
	}

	// ユーザー認証中。
	log.Debug("session " + mosaic(sess.Id()) + " is in authentication process")

	req := newSelectRequest(r)
	if sess.Ticket() == "" {
		// アカウント選択中でない。
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Access_denied, "not in interactive process", nil), sess)
	} else if req.ticket() != sess.Ticket() {
		// 無効なアカウント選択券。
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Access_denied, "invalid ticket "+mosaic(req.ticket()), nil), sess)
	}

	// チケットが有効だった。
	log.Debug("Ticket " + mosaic(req.ticket()) + " is OK")

	if req.accountName() == "" {
		// アカウント選択情報不備。
		log.Warn("Account is not specified")
		return sys.redirectToSelectUi(w, r, sess, "Please select your account")
	}

	// アカウント選択情報があった。
	log.Debug("Account " + req.accountName() + " is specified")

	acnt, err := sys.acntDb.GetByName(req.accountName())
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sess)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug("Specified accout " + req.accountName() + " was not found")
		return sys.redirectToSelectUi(w, r, sess, "Accout "+req.accountName()+" was not found. Please select your account")
	}

	// アカウント選択できた。
	log.Info("Specified account " + req.accountName() + " is exist")

	sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	if lang := req.language(); lang != "" {
		sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug("Language " + lang + " was selected")
	}

	return sys.afterSelect(w, r, sess, nil, acnt)
}

// アカウント選択が終わったところから。
func (sys *system) afterSelect(w http.ResponseWriter, r *http.Request, sess *session.Element, ta tadb.Element, acnt account.Element) error {

	prmpts := sess.Request().Prompt()
	if prmpts[prmptLogin] {
		if prmpts[prmptNone] {
			return sys.redirectError(w, r, newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil), sess)
		}

		return sys.redirectToLoginUi(w, r, sess, "Please log in")
	}

	if sess.Account() != nil && sess.Account().LoggedIn() {
		if maxAge := sess.Request().MaxAge(); maxAge < 0 || time.Now().Sub(sess.Account().LoginDate()) <= time.Duration(maxAge*time.Second) {
			// ログイン済み。
			return sys.afterLogin(w, r, sess, ta, acnt)
		}
	}

	if prmpts[prmptNone] {
		return sys.redirectError(w, r, newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil), sess)
	}

	return sys.redirectToLoginUi(w, r, sess, "Please log in")
}
