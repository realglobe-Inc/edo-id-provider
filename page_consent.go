package main

import (
	"fmt"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
	"time"
)

const (
	consHtml = "consent.html"
)

const (
	formClm    = "claim"
	formUsrNam = "username"
	formTaNam  = "client_name"
)

// 同意ページにリダイレクトする。
func redirectConsentUi(w http.ResponseWriter, r *http.Request, sys *system, sess *session, hint string) error {

	v := url.Values{}
	v.Set(formUsrNam, sess.currentAccountName())
	v.Set(formTaId, sess.request().ta())
	v.Set(formTaNam, sess.request().taName())
	scops := sess.request().scopes()
	if len(scops) > 0 {
		buff := valueSetToForm(scops)
		v.Set(formScop, buff)
	}
	clms := sess.request().claimNames()
	if len(clms) > 0 {
		buff := valueSetToForm(clms)
		v.Set(formClm, buff)
	}
	if hint != "" {
		v.Set(formHint, hint)
	}
	var query string
	if len(v) > 0 {
		query = "?" + v.Encode()
	}

	// 同意ページに渡すクエリパラメータを生成。

	tic, err := sys.newTicket()
	if err != nil {
		return redirectServerError(w, r, sys, sess, sess.request().redirectUri(), erro.Wrap(err))
	}
	sess.setConsentTicket(tic)

	// 同意券を発行。
	log.Debug("Consent ticket " + mosaic(tic) + " was generated")

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
	http.Redirect(w, r, sys.uiUri+"/"+consHtml+query+"#"+tic, http.StatusFound)
	return nil
}

// 同意ページからの入力を受け付けて続きをする。
func consentPage(w http.ResponseWriter, r *http.Request, sys *system) error {
	req := newConsentRequest(r)

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

	tic := sess.consentTicket()
	if tic == "" {
		// 同意中でない。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "not in consent process")
	} else if t := req.ticket(); t != tic {
		// 無効な同意券。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "invalid consent ticket "+mosaic(t))
	}

	// 同意券が有効だった。
	log.Debug("Consent ticket " + mosaic(tic) + " is OK")

	scops, clms, denyScops, denyClms := req.consentInfo()
	if scops == nil || clms == nil || denyScops == nil || denyClms == nil {
		// 同意情報不備。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, "consent info was not found")
	}

	// 同意情報があった。
	log.Debug("Consent info was found")

	sess.consent(scops, clms, denyScops, denyClms)
	if s, c := sess.unconsentedEssentials(); len(s) > 0 || len(c) > 0 {
		// 同意が足りなかった。
		return redirectError(w, r, sys, sess, authReq.redirectUri(), errAccDeny, fmt.Sprint("essential consent for ", s, c, " was denied"))
	}

	// 同意できた。
	log.Debug("Essential consent was given")

	return afterConsent(w, r, sys, sess)
}

func afterConsent(w http.ResponseWriter, r *http.Request, sys *system, sess *session) error {
	return publishCode(w, r, sys, sess)
}
