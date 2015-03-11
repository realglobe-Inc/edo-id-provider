package main

import (
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

// 認可コード発行。
func publishCode(w http.ResponseWriter, r *http.Request, sys *system, sess *session) error {

	authReq := sess.request() // commit すると消えるので取っとく。

	consScops, consClms, denyScops, denyClms := sess.commit()

	if !consScops[scopOpId] {
		// openid すら許されなかった。
		return redirectError(w, r, sys, sess, authReq, newIdpError(errAccDeny, "user denied openid", 0, nil))
	}

	codId, err := sys.codCont.newId()
	if err != nil {
		return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
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
		return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
	}

	// 認可コードを発行した。
	log.Debug("Code " + mosaic(cod.id()) + " was published")

	if sess.id() == "" {
		id, err := sys.sessCont.newId()
		if err != nil {
			return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
		}
		sess.setId(id)
	}
	sess.setExpirationDate(time.Now().Add(sys.sessExpiDur))
	if err := sys.sessCont.put(sess); err != nil {
		return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	if err := sys.consCont.put(sess.currentAccount(), authReq.ta(), consScops, consClms, denyScops, denyClms); err != nil {
		return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
	}

	// 認可コードを IdP の ID を含んだ JWS にする。
	jt := jwt.New()
	jt.SetHeader(jwtAlg, algNone)
	jt.SetClaim(clmJti, cod.id())
	jt.SetClaim(clmIss, sys.selfId)
	buff, err := jt.Encode(nil, nil)
	if err != nil {
		return redirectError(w, r, sys, sess, authReq, erro.Wrap(err))
	}
	encCod := string(buff)

	q := authReq.redirectUri().Query()
	q.Set(formCod, encCod)
	if stat := authReq.state(); stat != "" {
		q.Set(formStat, stat)
	}
	authReq.redirectUri().RawQuery = q.Encode()
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Path:     "/",
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	http.Redirect(w, r, authReq.redirectUri().String(), http.StatusFound)
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

	if req.rawRequest() != "" {
		if err := req.parseRequest(t.keys(), map[string]interface{}{sys.sigKid: sys.sigKey}); err != nil {
			return newIdpError(errInvReq, erro.Unwrap(err).Error(), http.StatusBadRequest, erro.Wrap(err))
		}
	}

	if req.rawRedirectUri() == "" {
		return newIdpError(errInvReq, "no "+formRediUri, http.StatusBadRequest, nil)
	} else if !t.redirectUris()[req.rawRedirectUri()] {
		return newIdpError(errInvReq, formRediUri+" "+req.rawRedirectUri()+" is not registered", http.StatusBadRequest, nil)
	} else if err := req.parseRedirectUri(); err != nil {
		return newIdpError(errInvReq, erro.Unwrap(err).Error(), http.StatusBadRequest, erro.Wrap(err))
	}

	// リダイレクト先には問題無い。
	log.Debug("Redirect URI " + req.rawRedirectUri() + " is OK")

	if !req.scopes()[scopOpId] {
		return redirectError(w, r, sys, nil, req, newIdpError(errInvReq, formScop+" has no "+scopOpId, 0, nil))
	}

	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return redirectError(w, r, sys, nil, req, newIdpError(errInvReq, k+" is overlapped", 0, nil))
		}
	}

	// scope には問題無い。
	log.Debug("Scope has " + scopOpId)

	if l := len(req.responseType()); l == 0 {
		return redirectError(w, r, sys, nil, req, newIdpError(errInvReq, "no "+formRespType, 0, nil))
	} else if l != 1 || !req.responseType()[respTypeCod] {
		return redirectError(w, r, sys, nil, req, newIdpError(errUnsuppRespType, formRespType+" is not "+respTypeCod, 0, nil))
	}

	// response_type には問題無い。
	log.Debug("Response type is " + respTypeCod)

	if err := req.parse(); err != nil {
		return redirectError(w, r, sys, nil, req, newIdpError(errInvReq, erro.Unwrap(err).Error(), 0, erro.Wrap(err)))
	}

	// リクエストの文法には問題無い。
	log.Debug("Authentication request is OK")

	var sess *session
	if sessId := newBrowserRequest(r).session(); sessId != "" {
		// セッションが通知された。
		log.Debug("Session " + mosaic(sessId) + " is declared")

		var err error
		sess, err = sys.sessCont.get(sessId)
		if err != nil {
			return redirectError(w, r, sys, nil, req, erro.Wrap(err))
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

	if req.prompts()[prmptSelAcc] {
		if req.prompts()[prmptNone] {
			return redirectError(w, r, sys, nil, req, newIdpError(errAccSelReq, "cannot select account without UI", 0, nil))
		}

		return redirectSelectUi(w, r, sys, sess, "")
	}

	return afterSelect(w, r, sys, sess)
}
