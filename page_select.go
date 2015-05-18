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
	"github.com/realglobe-Inc/edo-id-provider/request"
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
	return request.ValuesForm(a)
}

// アカウント選択 UI にリダイレクトする。
func (sys *system) redirectToSelectUi(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(sys.pathSelUi)
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
	}

	// アカウント選択ページに渡すクエリパラメータを生成。
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

	log.Info(sender, ": Redirect to select UI")
	return sys.redirectTo(w, r, uri, sender, sess)
}

// アカウント UI ページからの入力を受け付けて続きをする。
func (sys *system) selectPage(w http.ResponseWriter, r *http.Request) (err error) {
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
		// セッションを新規発行。
		sess = session.New(randomString(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info(sender, ": Generated new session "+mosaic(sess.Id())+" but not yet saved")
	}

	// セッションは決まった。

	authReq := sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return sys.returnError(w, r, erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil)), sender, sess)
	}

	// ユーザー認証中。
	log.Debug(sender, ": Session is in authentication process")

	req := newSelectRequest(r)
	if sess.Ticket() == "" {
		// アカウント選択中でない。
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Access_denied, "not in interactive process", nil)), sender, sess)
	} else if req.ticket() != sess.Ticket() {
		// 無効なアカウント選択券。
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Access_denied, "invalid ticket "+mosaic(req.ticket()), nil)), sender, sess)
	}

	// チケットが有効だった。
	log.Debug(sender, ": Ticket "+mosaic(req.ticket())+" is OK")

	if req.accountName() == "" {
		// アカウント選択情報不備。
		log.Warn(sender, ": Account is not specified")
		return sys.redirectToSelectUi(w, r, sender, sess, "Please select your account")
	}

	// アカウント選択情報があった。
	log.Debug(sender, ": Account "+req.accountName()+" is specified")

	acnt, err := sys.acntDb.GetByName(req.accountName())
	if err != nil {
		return sys.redirectError(w, r, erro.Wrap(err), sender, sess)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug(sender, ": Specified accout "+req.accountName()+" is not exist")
		return sys.redirectToSelectUi(w, r, sender, sess, "Accout "+req.accountName()+" is not exist. Please select your account")
	}

	// アカウント選択できた。
	log.Info(sender, ": Specified account "+req.accountName()+" is exist")

	sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	if lang := req.language(); lang != "" {
		sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug(sender, ": Language "+lang+" was selected")
	}

	return sys.afterSelect(w, r, sender, sess, nil, acnt)
}

// アカウント選択が終わったところから。
func (sys *system) afterSelect(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, ta tadb.Element, acnt account.Element) error {

	prmpts := sess.Request().Prompt()
	if prmpts[tagLogin] {
		if prmpts[tagNone] {
			return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil)), sender, sess)
		}

		return sys.redirectToLoginUi(w, r, sender, sess, "Please log in")
	}

	if sess.Account() != nil && sess.Account().LoggedIn() {
		if maxAge := sess.Request().MaxAge(); maxAge < 0 || time.Now().Sub(sess.Account().LoginDate()) <= time.Duration(maxAge*time.Second) {
			// ログイン済み。
			return sys.afterLogin(w, r, sender, sess, ta, acnt)
		}
	}

	if prmpts[tagNone] {
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil)), sender, sess)
	}

	return sys.redirectToLoginUi(w, r, sender, sess, "Please log in")
}
