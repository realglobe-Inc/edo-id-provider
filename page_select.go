package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"time"
)

// アカウント選択ページにリダイレクトする。
func redirectSelectUi(w http.ResponseWriter, r *http.Request, sys *system, sess *session, hint string) error {
	// TODO 試行回数でエラー。

	v := url.Values{}
	if accs := sess.accountNames(); len(accs) > 0 {
		buff, err := json.Marshal(strset.StringSet(accs))
		if err != nil {
			return redirectError(w, r, sys, sess, sess.request(), erro.Wrap(err))
		}

		v.Set(formAccNames, string(buff))
	}
	if disp := sess.request().display(); disp != "" {
		v.Set(formDisp, disp)
	}
	if loc, locs := sess.locale(), sess.request().uiLocales(); loc != "" || len(locs) > 0 {
		v.Set(formLocs, valuesToForm(append([]string{loc}, locs...)))
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
		return redirectError(w, r, sys, sess, sess.request(), erro.Wrap(err))
	}
	sess.setSelectTicket(tic)

	// アカウント選択券を発行。
	log.Debug("Select ticket " + mosaic(tic) + " was generated")

	if sess.id() == "" {
		id, err := sys.sessCont.newId()
		if err != nil {
			return redirectError(w, r, sys, sess, sess.request(), erro.Wrap(err))
		}
		sess.setId(id)
	}
	sess.setExpirationDate(time.Now().Add(sys.sessExpiDur))
	if err := sys.sessCont.put(sess); err != nil {
		return redirectError(w, r, sys, sess, sess.request(), erro.Wrap(err))
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
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	http.Redirect(w, r, sys.uiUri+"/"+selHtml+query+"#"+tic, http.StatusFound)
	return nil

}

// アカウント選択ページからの入力を受け付けて続きをする。
func selectPage(w http.ResponseWriter, r *http.Request, sys *system) error {
	req := newSelectRequest(r)

	sessId := req.session()
	if sessId == "" {
		// セッションが通知されてない。
		return newIdpError(errInvReq, "no session", http.StatusBadRequest, nil)
	}

	// セッションが通知された。
	log.Debug("session " + mosaic(sessId) + " is declared")

	sess, err := sys.sessCont.get(sessId)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		// セッションなんて無かった。
		return newIdpError(errInvReq, "no session "+mosaic(sessId), http.StatusBadRequest, nil)
	} else if !sess.valid() {
		// 無効なセッション。
		return newIdpError(errInvReq, "invalid session "+mosaic(sessId), http.StatusBadRequest, nil)
	}

	// セッションが有効だった。
	log.Debug("session " + mosaic(sessId) + " is exist")

	authReq := sess.request()
	if authReq == nil {
		// ユーザー認証・認可処理が始まっていない。
		return newIdpError(errInvReq, "session "+mosaic(sessId)+" is not in authentication process", http.StatusBadRequest, nil)
	}

	// ユーザー認証・認可処理中。
	log.Debug("session " + mosaic(sessId) + " is in authentication process")

	tic := sess.selectTicket()
	if tic == "" {
		// アカウント選択中でない。
		return redirectError(w, r, sys, sess, authReq, newIdpError(errAccDeny, "not in account selection process", 0, nil))
	} else if t := req.ticket(); t != tic {
		// 無効なアカウント選択券。
		return redirectError(w, r, sys, sess, authReq, newIdpError(errAccDeny, "invalid account selection ticket "+mosaic(t), 0, nil))
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
		return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
	} else if acc == nil {
		// アカウントが無い。
		log.Debug("Accout " + accName + " was not found")
		return redirectSelectUi(w, r, sys, sess, "accout "+accName+" was not found")
	}

	// アカウント選択できた。
	log.Debug("Account " + accName + " was selected")

	sess.selectAccount(acc)

	if loc := req.locale(); loc != "" {
		sess.setLocale(loc)

		// 言語を選択してた。
		log.Debug("Locale " + loc + " was selected")
	}

	return afterSelect(w, r, sys, sess)
}

// アカウント選択が無事終わった後の処理。
func afterSelect(w http.ResponseWriter, r *http.Request, sys *system, sess *session) error {

	prmpts := sess.request().prompts()
	if prmpts[prmptLogin] {
		if prmpts[prmptNone] {
			return redirectError(w, r, sys, sess, sess.request(), newIdpError(errLoginReq, "cannot login without UI", 0, nil))
		}

		log.Debug("Login is forced")
		return redirectLoginUi(w, r, sys, sess, "")
	}

	if sess.currentAccountAuthenticated() && (sess.request().maxAge() < 0 ||
		time.Now().Sub(sess.currentAccountDate()) <= time.Duration(sess.request().maxAge())*time.Second) {
		// ログイン済み。
		log.Debug("Logged in")
		return afterLogin(w, r, sys, sess)
	}

	log.Debug("Logged is required")

	if prmpts[prmptNone] {
		return redirectError(w, r, sys, sess, sess.request(), newIdpError(errLoginReq, "cannot login without UI", 0, nil))
	}

	return redirectLoginUi(w, r, sys, sess, "")
}
