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
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-idp-selector/ticket"
	jsonutil "github.com/realglobe-Inc/edo-lib/json"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"net/url"
	"time"
)

// アカウント UI ページからの入力を受け付けて続きをする。
func (this *Page) HandleSelect(w http.ResponseWriter, r *http.Request) {
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

	log.Info(logPref, "Received select request")
	defer log.Info(logPref, "Handled select request")

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
		// セッションを新規発行。
		sess = session.New(this.idGen.String(this.sessLen), now.Add(this.sessExpIn))
		log.Info(logPref, "Generated new session "+logutil.Mosaic(sess.Id())+" but not yet saved")
	}

	// セッションは決まった。

	env := (&environment{this, logPref, sess})
	if err := env.selectServe(w, r); err != nil {
		env.respondErrorHtml(w, r, erro.Wrap(err))
		return
	}
}

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
func (this *environment) redirectToSelectUi(w http.ResponseWriter, r *http.Request, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(this.pathSelUi)
	if err != nil {
		return erro.Wrap(err)
	}

	// アカウント選択ページに渡すクエリパラメータを生成。
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

	log.Info(this.logPref, "Redirect to select UI")
	this.redirectTo(w, r, uri)
	return nil
}

func (this *environment) selectServe(w http.ResponseWriter, r *http.Request) error {
	authReq := this.sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil))
	}

	// ユーザー認証中。
	log.Debug(this.logPref, "Session is in authentication process")

	req, err := parseSelectRequest(r)
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

	if req.accountName() == "" {
		// アカウント選択情報不備。
		log.Warn(this.logPref, "Account is not specified")
		return this.redirectToSelectUi(w, r, "Please select your account")
	}

	// アカウント選択情報があった。
	log.Debug(this.logPref, "Account "+req.accountName()+" is specified")

	acnt, err := this.acntDb.GetByName(req.accountName())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug(this.logPref, "Specified accout "+req.accountName()+" is not exist")
		return this.redirectToSelectUi(w, r, "Accout "+req.accountName()+" is not exist. Please select your account")
	}

	// アカウント選択できた。
	log.Info(this.logPref, "Specified account "+req.accountName()+" is exist")

	if cur := this.sess.Account(); cur == nil || cur.Id() != acnt.Id() {
		this.sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	}

	if lang := req.language(); lang != "" {
		this.sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug(this.logPref, "Language "+lang+" was selected")
	}

	return this.afterSelect(w, r, nil, acnt)
}

// アカウント選択が終わったところから。
func (this *environment) afterSelect(w http.ResponseWriter, r *http.Request, ta tadb.Element, acnt account.Element) error {

	prmpts := this.sess.Request().Prompt()
	if prmpts[tagLogin] {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil))
		}

		return this.redirectToLoginUi(w, r, "Please log in")
	}

	if this.sess.Account() != nil && this.sess.Account().LoggedIn() {
		if maxAge := this.sess.Request().MaxAge(); maxAge < 0 || time.Now().Sub(this.sess.Account().LoginDate()) <= time.Duration(maxAge*time.Second) {
			// ログイン済み。
			return this.afterLogin(w, r, ta, acnt)
		}
	}

	if prmpts[tagNone] {
		return erro.Wrap(newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil))
	}

	return this.redirectToLoginUi(w, r, "Please log in")
}
