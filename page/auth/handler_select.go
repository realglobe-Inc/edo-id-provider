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
	var sender *request.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondHtml(w, r, erro.New(rcv), this.errTmpl, sender)
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

	sender = request.Parse(r, this.sessLabel)
	log.Info(sender, ": Received select request")
	defer log.Info(sender, ": Handled select request")

	if err := this.selectServe(w, r, sender); err != nil {
		idperr.RespondHtml(w, r, erro.Wrap(err), this.errTmpl, sender)
		return
	}
	return
}

func (this *Page) selectServe(w http.ResponseWriter, r *http.Request, sender *request.Request) (err error) {
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
		// セッションを新規発行。
		sess = session.New(this.idGen.String(this.sessLen), now.Add(this.sessExpIn))
		log.Info(sender, ": Generated new session "+logutil.Mosaic(sess.Id())+" but not yet saved")
	}

	// セッションは決まった。

	if err := this.selectServeWithSession(w, r, sender, sess); err != nil {
		return this.respondErrorHtml(w, r, erro.Wrap(err), sender, sess)
	}
	return nil
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
func (this *Page) redirectToSelectUi(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, msg string) error {
	// TODO 試行回数でエラー。

	uri, err := url.Parse(this.pathSelUi)
	if err != nil {
		return this.respondErrorHtml(w, r, erro.Wrap(err), sender, sess)
	}

	// アカウント選択ページに渡すクエリパラメータを生成。
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

	log.Info(sender, ": Redirect to select UI")
	return this.redirectTo(w, r, uri, sender, sess)
}

func (this *Page) selectServeWithSession(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element) error {
	authReq := sess.Request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return erro.Wrap(idperr.New(idperr.Invalid_request, "session is not in authentication process", http.StatusBadRequest, nil))
	}

	// ユーザー認証中。
	log.Debug(sender, ": Session is in authentication process")

	req, err := parseSelectRequest(r)
	if err != nil {
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, erro.Unwrap(err).Error(), err))
	} else if req.ticket() != sess.Ticket() {
		// 無効なアカウント選択券。
		return erro.Wrap(newErrorForRedirect(idperr.Access_denied, "invalid ticket "+logutil.Mosaic(req.ticket()), nil))
	}

	// チケットが有効だった。
	log.Debug(sender, ": Ticket "+logutil.Mosaic(req.ticket())+" is OK")

	if req.accountName() == "" {
		// アカウント選択情報不備。
		log.Warn(sender, ": Account is not specified")
		return this.redirectToSelectUi(w, r, sender, sess, "Please select your account")
	}

	// アカウント選択情報があった。
	log.Debug(sender, ": Account "+req.accountName()+" is specified")

	acnt, err := this.acntDb.GetByName(req.accountName())
	if err != nil {
		return erro.Wrap(err)
	} else if acnt == nil {
		// アカウントが無い。
		log.Debug(sender, ": Specified accout "+req.accountName()+" is not exist")
		return this.redirectToSelectUi(w, r, sender, sess, "Accout "+req.accountName()+" is not exist. Please select your account")
	}

	// アカウント選択できた。
	log.Info(sender, ": Specified account "+req.accountName()+" is exist")

	if cur := sess.Account(); cur == nil || cur.Id() != acnt.Id() {
		sess.SelectAccount(session.NewAccount(acnt.Id(), acnt.Name()))
	}

	if lang := req.language(); lang != "" {
		sess.SetLanguage(lang)

		// 言語を選択してた。
		log.Debug(sender, ": Language "+lang+" was selected")
	}

	return this.afterSelect(w, r, sender, sess, nil, acnt)
}

// アカウント選択が終わったところから。
func (this *Page) afterSelect(w http.ResponseWriter, r *http.Request, sender *request.Request, sess *session.Element, ta tadb.Element, acnt account.Element) error {

	prmpts := sess.Request().Prompt()
	if prmpts[tagLogin] {
		if prmpts[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil))
		}

		return this.redirectToLoginUi(w, r, sender, sess, "Please log in")
	}

	if sess.Account() != nil && sess.Account().LoggedIn() {
		if maxAge := sess.Request().MaxAge(); maxAge < 0 || time.Now().Sub(sess.Account().LoginDate()) <= time.Duration(maxAge*time.Second) {
			// ログイン済み。
			return this.afterLogin(w, r, sender, sess, ta, acnt)
		}
	}

	if prmpts[tagNone] {
		return erro.Wrap(newErrorForRedirect(idperr.Login_required, "cannot login without UI", nil))
	}

	return this.redirectToLoginUi(w, r, sender, sess, "Please log in")
}
