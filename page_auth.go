package main

import (
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
	"time"
)

// 認可コード発行。
func publishCode(w http.ResponseWriter, r *http.Request, sys *system, sess *session) error {

	authReq := sess.request() // commit すると消えるので取っとく。

	consScops, consClms, denyScops, denyClms := sess.commit()

	codId, err := sys.codCont.newId()
	if err != nil {
		return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	}
	cod := newCode(
		codId,
		sess.currentAccount(),
		authReq.ta(),
		authReq.redirectUri().String(),
		time.Now().Add(sys.codExpiDur),
		sys.tokExpiDur,
		consScops,
		consClms,
		authReq.nonce(),
		sess.currentAccountDate(),
	)
	log.Debug("Code " + mosaic(cod.id()) + " was generated.")

	if err := sys.codCont.put(cod); err != nil {
		return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	}

	// 認可コードを発行した。
	log.Debug("Code " + mosaic(cod.id()) + " was published")

	if sess.id() == "" {
		id, err := sys.sessCont.newId()
		if err != nil {
			return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
		}
		sess.setId(id)
	}
	sess.setExpirationDate(time.Now().Add(sys.sessExpiDur))
	if err := sys.sessCont.put(sess); err != nil {
		return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	if err := sys.consCont.put(sess.currentAccount(), authReq.ta(), consScops, consClms, denyScops, denyClms); err != nil {
		return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	}

	// 認可コードを IdP の ID を含んだ JWS にする。
	jws := util.NewJws()
	jws.SetHeader(jwtAlg, algNone)
	jws.SetClaim(clmJti, cod.id())
	jws.SetClaim(clmIss, sys.selfId)
	if err := jws.Sign(nil); err != nil {
		return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	}
	buff, err := jws.Encode()
	if err != nil {
		return redirectError(w, r, sys, sess, authReq.redirectUri(), erro.Wrap(err))
	}
	encCod := string(buff)

	rediUri := authReq.redirectUri()
	q := rediUri.Query()
	q.Set(formCod, encCod)
	if stat := authReq.state(); stat != "" {
		q.Set(formStat, stat)
	}
	rediUri.RawQuery = q.Encode()
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Path:     "/",
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r, rediUri.String(), http.StatusFound)
	return nil
}

// ユーザー認証・認可開始ページ。
func authPage(w http.ResponseWriter, r *http.Request, sys *system) error {
	req, err := newAuthRequest(r)
	if err != nil {
		return erro.Wrap(err)
	}

	if req.ta() == "" {
		return newIdpError(errInvReq, "no "+formTaId, http.StatusBadRequest, nil)
	}

	// TA が指定されてる。
	log.Debug("TA " + req.ta() + " is declared")

	t, err := sys.taCont.get(req.ta())
	if err != nil {
		return erro.Wrap(err)
	} else if t == nil {
		return newIdpError(errInvReq, "invalid TA "+req.ta(), http.StatusBadRequest, nil)
	}

	// TA は存在する。
	log.Debug("TA " + t.id() + " is exist")
	req.setTaName(t.name())

	if req.rawRedirectUri() == "" {
		return newIdpError(errInvReq, "no "+formRediUri, http.StatusBadRequest, nil)
	} else if !t.redirectUris()[req.rawRedirectUri()] {
		return newIdpError(errInvReq, formRediUri+" "+req.rawRedirectUri()+" is not registered", http.StatusBadRequest, nil)
	}
	rediUri, err := url.Parse(req.rawRedirectUri())
	if err != nil {
		return newIdpError(errInvReq, erro.Unwrap(err).Error(), http.StatusBadRequest, erro.Wrap(err))
	}

	// リダイレクト先には問題無い。
	log.Debug("Redirect URI " + req.rawRedirectUri() + " is OK")
	req.setRedirectUri(rediUri)

	if !req.scopes()[scopOpId] {
		return redirectError(w, r, sys, nil, req.redirectUri(), newIdpError(errInvScop, formScop+" has no "+scopOpId, 0, nil))
	}

	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return redirectError(w, r, sys, nil, req.redirectUri(), newIdpError(errInvReq, k+" is overlapped", 0, nil))
		}
	}

	// scope には問題無い。
	log.Debug("Scope has " + scopOpId)

	if len(req.responseType()) != 1 || !req.responseType()[respTypeCod] {
		return redirectError(w, r, sys, nil, req.redirectUri(), newIdpError(errUnsuppRespType, formRespType+" is not "+respTypeCod, 0, nil))
	}

	// response_type には問題無い。
	log.Debug("Response type is " + respTypeCod)

	var sess *session
	if sessId := newBrowserRequest(r).session(); sessId != "" {
		// セッションが通知された。
		log.Debug("Session " + mosaic(sessId) + " is declared")

		var err error
		sess, err = sys.sessCont.get(sessId)
		if err != nil {
			return redirectError(w, r, sys, nil, req.redirectUri(), erro.Wrap(err))
		} else if sess == nil {
			// セッションなんて無かった。
			log.Warn("Session " + mosaic(sessId) + " is not exist")
		} else {
			// セッションがあった。
			log.Debug("Session " + mosaic(sessId) + " is exist")
		}
	}

	if sess == nil {
		sess = newSession()
		log.Debug("New session was generated but not yet registered")
	}
	sess.startRequest(req)

	prmpts := req.prompts()
	if prmpts[prmptSelAcc] {
		if prmpts[prmptNone] {
			return redirectError(w, r, sys, nil, req.redirectUri(), newIdpError(errAccSelReq, "cannot select account without UI", 0, nil))
		}

		return redirectSelectUi(w, r, sys, sess, "")
	}

	return afterSelect(w, r, sys, sess)
}
