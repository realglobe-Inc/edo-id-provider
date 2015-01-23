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
	loginHtml = "login.html"
)

const (
	formHint     = "hint"
	formUsrNams  = "usernames"
	formLoginTic = "ticket"
)

// ログインページにリダイレクトする。
func redirectLoginUi(w http.ResponseWriter, r *http.Request, sys *system, sess *session, hint string) error {

	// TODO 試行回数でエラー。

	v := url.Values{}
	if accNames := sess.accountNames(); len(accNames) > 0 {
		buff, err := json.Marshal(util.StringSet(accNames))
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

	// ログインページに渡すクエリパラメータを生成。

	tic, err := sys.newTicket()
	if err != nil {
		return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
	}
	sess.setLoginTicket(tic)

	// ログイン券を発行。
	log.Debug("Login ticket " + mosaic(tic) + " was generated")

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
	http.Redirect(w, r, sys.uiUri+"/"+loginHtml+query+"#"+tic, http.StatusFound)
	return nil
}

// ログインページからの入力を受け付けて続きをする。
func loginPage(w http.ResponseWriter, r *http.Request, sys *system) error {
	req := newLoginRequest(r)

	sessId := req.session()
	if sessId == "" {
		// セッションが通知されてない。
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no session", nil))
	}

	// セッションが通知された。
	log.Debug("session " + mosaic(sessId) + " is declared")

	sess, err := sys.sessCont.get(sessId)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		// セッションなんて無かった。
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no session "+mosaic(sessId), nil))
	} else if !sess.valid() {
		// 無効なセッション。
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "invalid session "+mosaic(sessId), nil))
	}

	// セッションが有効だった。
	log.Debug("session " + mosaic(sessId) + " is exist")

	authReq := sess.request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "session "+mosaic(sessId)+" is not in authentication process", nil))
	}

	// ユーザー認証・認可処理中。
	log.Debug("session " + mosaic(sessId) + " is in authentication process")

	tic := sess.loginTicket()
	if tic == "" {
		// ログイン中でない。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "not in login process")
	} else if t := req.ticket(); t != tic {
		// 無効なログイン券。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "invalid login ticket "+mosaic(t))
	}

	// ログイン券が有効だった。
	log.Debug("Login ticket " + mosaic(tic) + " is OK")

	accName, passwd := req.loginInfo()
	if accName == "" || passwd == "" {
		// ログイン情報不備。
		log.Debug("Login info was not found")
		return redirectLoginUi(w, r, sys, sess, "login info was not found")
	}

	// ログイン情報があった。
	log.Debug("Login info was found")

	acc, err := sys.accCont.getByName(accName)
	if err != nil {
		return redirectServerError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	} else if acc == nil {
		// アカウントが無い。
		log.Debug("Accout " + accName + " was not found")
		return redirectLoginUi(w, r, sys, sess, "accout "+accName+" was not found")
	}
	if passwd != acc.password() {
		// パスワード間違い。
		log.Debug("Password differs from " + accName + "'s password")
		return redirectLoginUi(w, r, sys, sess, "password differs from "+accName+"'s password")
	}

	// ログインできた。
	log.Debug("Account " + accName + " (" + acc.id() + ") was authenticated")

	sess.loginAccount(acc)

	return afterLogin(w, r, sys, sess)
}

// ログインが無事終わった後の処理。
func afterLogin(w http.ResponseWriter, r *http.Request, sys *system, sess *session) error {

	prmpts := sess.request().prompts()
	if prmpts[prmptCons] && prmpts[prmptNone] {
		return redirectError(w, r, sys, sess, sess.request().redirectUri(), errConsReq, "cannot consent without UI")
	}

	if prmpts[prmptCons] {
		log.Debug("Consent is forced")
		return redirectConsentUi(w, r, sys, sess, "")
	}

	// TODO essential クレームの有無。
	// TODO value, values 指定クレームの検査。

	// 事前同意を調べる。
	scops, clms, err := sys.consCont.get(sess.currentAccount(), sess.request().ta())
	if err != nil {
		return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
	}
	if satisfiable(scops, clms, sess.request().scopes(), sess.request().claimNames()) {
		// 事前同意で十分。
		log.Debug("Preliminarily consented")
		return afterConsent(w, r, sys, sess)
	}

	log.Debug("Consent is required")

	if prmpts[prmptNone] {
		return redirectError(w, r, sys, sess, sess.request().redirectUri(), errConsReq, "cannot consent without UI")
	}

	return redirectConsentUi(w, r, sys, sess, "")
}
