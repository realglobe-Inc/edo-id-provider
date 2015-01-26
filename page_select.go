package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
	"time"
)

const (
	selHtml = "select.html"
)

// アカウント選択ページにリダイレクトする。
func redirectSelectUi(w http.ResponseWriter, r *http.Request, sys *system, sess *session, hint string) error {
	// TODO 試行回数でエラー。

	v := url.Values{}
	if accs := sess.accountNames(); len(accs) > 0 {
		buff, err := json.Marshal(util.StringSet(accs))
		if err != nil {
			return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
		}

		v.Set(formUsrNams, string(buff))
	}
	if hint != "" {
		v.Set(formHint, hint)
	}
	var query string
	if len(v) > 0 {
		query = "?" + v.Encode()
	}

	// アカウント選択ページに渡すクエリパラメータを生成。

	tic, err := sys.newTicket()
	if err != nil {
		return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
	}
	sess.setSelectTicket(tic)

	// アカウント選択券を発行。
	log.Debug("Select ticket " + mosaic(tic) + " was generated")

	if sess.id() == "" {
		id, err := sys.sessCont.newId()
		if err != nil {
			return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
		}
		sess.setId(id)
	}
	sess.setExpirationDate(time.Now().Add(sys.sessExpiDur))
	if err := sys.sessCont.put(sess); err != nil {
		return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
	}

	// セッションを保存した。
	log.Debug("Session " + mosaic(sess.id()) + " was saved")

	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Path:     "/",
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r, sys.uiUri+"/"+selHtml+query+"#"+tic, http.StatusFound)
	return nil

}

// アカウント選択ページからの入力を受け付けて続きをする。
func selectPage(w http.ResponseWriter, r *http.Request, sys *system) error {
	req := newSelectRequest(r)

	sessId := req.session()
	if sessId == "" {
		// セッションが通知されてない。
		return responseError(w, http.StatusBadRequest, errInvReq, "no session")
	}

	// セッションが通知された。
	log.Debug("session " + mosaic(sessId) + " is declared")

	sess, err := sys.sessCont.get(sessId)
	if err != nil {
		return responseServerError(w, http.StatusInternalServerError, erro.Wrap(err))
	} else if sess == nil {
		// セッションなんて無かった。
		return responseError(w, http.StatusBadRequest, errInvReq, "no session "+mosaic(sessId))
	} else if !sess.valid() {
		// 無効なセッション。
		return responseError(w, http.StatusBadRequest, errInvReq, "invalid session "+mosaic(sessId))
	}

	// セッションが有効だった。
	log.Debug("session " + mosaic(sessId) + " is exist")

	authReq := sess.request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return responseError(w, http.StatusBadRequest, errInvReq, "session "+mosaic(sessId)+" is not in authentication process")
	}

	// ユーザー認証・認可処理中。
	log.Debug("session " + mosaic(sessId) + " is in authentication process")

	tic := sess.selectTicket()
	if tic == "" {
		// アカウント選択中でない。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "not in account selection process")
	} else if t := req.ticket(); t != tic {
		// 無効なアカウント選択券。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "invalid account selection ticket "+mosaic(t))
	}

	// アカウント選択券が有効だった。
	log.Debug("Account selection ticket " + mosaic(tic) + " is OK")

	accName := req.selectInfo()
	if accName == "" {
		// アカウント選択情報不備。
		log.Debug("Account selection info was not found")
		return redirectSelectUi(w, r, sys, sess, "account selection info was not found")
	}

	// アカウント選択情報があった。
	log.Debug("Account selection info was found")

	acc, err := sys.accCont.getByName(accName)
	if err != nil {
		return redirectServerError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	} else if acc == nil {
		// アカウントが無い。
		log.Debug("Accout " + accName + " was not found")
		return redirectSelectUi(w, r, sys, sess, "accout "+accName+" was not found")
	}

	// アカウント選択できた。
	log.Debug("Account " + accName + " was selected")

	sess.selectAccount(acc)

	return afterSelect(w, r, sys, sess)
}

// アカウント選択が無事終わった後の処理。
func afterSelect(w http.ResponseWriter, r *http.Request, sys *system, sess *session) error {

	prmpts := sess.request().prompts()
	if prmpts[prmptLogin] && prmpts[prmptNone] {
		return redirectError(w, r, sys, sess, sess.request().redirectUri(), errConsReq, "cannot login without UI")
	}

	if prmpts[prmptLogin] {
		log.Debug("Login is forced")
		return redirectLoginUi(w, r, sys, sess, "")
	}

	if sess.currentAccountAuthenticated() {
		// ログイン済み。
		log.Debug("Logged in")
		return afterLogin(w, r, sys, sess)
	}

	log.Debug("Logged is required")

	if prmpts[prmptNone] {
		return redirectError(w, r, sys, sess, sess.request().redirectUri(), errConsReq, "cannot login without UI")
	}

	return redirectLoginUi(w, r, sys, sess, "")
}
