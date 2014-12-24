package main

import (
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
)

// 認証ページ。
func authPage(sys *system, w http.ResponseWriter, r *http.Request) error {

	taId := getTaId(r)
	if taId == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formTaId, nil))
	}

	// TA が指定されてる。
	log.Debug("TA " + taId + " is declared")

	t, err := sys.taCont.get(taId)
	if err != nil {
		return erro.Wrap(err)
	} else if t == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "invalid TA "+taId, nil))
	}

	// TA は存在する。
	log.Debug("TA " + taId + " is exist")

	rediUriStr := getRedirectUri(r)
	if rediUriStr == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formRediUri, nil))
	} else if !t.hasRedirectUri(rediUriStr) {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, formRediUri+" "+rediUriStr+" is not registered", nil))
	}
	rediUri, err := url.Parse(rediUriStr)
	if err != nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, formRediUri+" "+rediUriStr+" is invalid URI", nil))
	}

	// リダイレクト先には問題無い。
	log.Debug("Redirect URI " + rediUriStr + " is OK")

	req := newAuthenticationRequest(r, t, rediUri)

	if !req.scopes()[scopOpId] {
		return redirectError(w, req, errInvScop, formScop+" has no "+scopOpId)
	}

	// scope には問題無い。
	log.Debug("Scope has " + scopOpId)

	if req.responseType() != respTypeCod {
		return redirectError(w, req, errUnsuppRespType, formRespType+" is not "+respTypeCod)
	}

	// response_type には問題無い。
	log.Debug("Response type is " + respTypeCod)

	return auth(sys, w, req)
}

func auth(sys *system, w http.ResponseWriter, r *authenticationRequest) error {
	// アカウント選択するかどうかまで。

	var sess *session
	if sessId := r.sessionId(); sessId != "" {
		// セッションが通知された。
		log.Debug("Session " + mosaic(sessId) + " is declared")

		var err error
		sess, err = sys.sessCont.get(sessId)
		if err != nil {
			return redirectServerError(w, r, erro.Wrap(err))
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

	if !r.prompts()[prmptSelAcc] {
		// アカウント選択は必要無い。
		log.Debug("Account selection is not required")

		return accountSelected(sys, w, r, sess)
	}

	// アカウント選択が指示されていた。
	log.Debug("Account selection is required")

	if r.prompts()[prmptNone] {
		return redirectError(w, r, errAccSelReq, "cannot select account without UI")
	}
	return accountSelect(sys, w, r, sess)
}

func accountSelected(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	// アカウント選択後。

	if r.prompts()[prmptLogin] {
		// 強制ユーザー認証。
		log.Debug("User authentication is forced")
		return unauthenticated(sys, w, r, sess)
	} else if !sess.isAuthenticated() {
		// ユーザー認証が必要。
		log.Debug("User authentication is required")
		return unauthenticated(sys, w, r, sess)
	} else {
		// ユーザー認証済み。
		log.Debug("User " + sess.account() + " is authenticated")
		return authenticated(sys, w, r, sess)
	}
}

func accountSelect(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	selCod := r.selectionCode()
	if selCod == "" {
		// アカウント選択 UI 表示前。
		log.Debug("Account selection starts")
		return redirectAccountSelectionUi(sys, w, r, sess)
	}

	if selCod != sess.selectionCode() {
		return redirectError(w, r, errAccDeny, "invalid account selection code")
	} else {
		// アカウント選択 UI 表示後だった。
		log.Debug("Account selection returned")

		accName, _ := r.authenticationData()
		if accName == "" {
			return redirectError(w, r, errAccDeny, "account was not selected")
		} else if sess.selectAccountByName(accName) {
			// アカウント選択できた。
			log.Debug("Account " + accName + " is selected")
		} else {
			// 認証済みでないアカウントを選択した。
			log.Debug("Maybe new login as " + accName)
		}

		return accountSelected(sys, w, r, sess)
	}
}

func authenticated(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	// ユーザー認証後。セッションはある。

	clms := r.claims()
	if len(clms) == 0 {
		// 必要な同意は無い。
		log.Debug("No consent is required")
		return publishCode(sys, w, r, sess)
	}
	t := r.ta()
	prmpts := r.prompts()

	// どのクレームが要求されてるか分かった。
	log.Debug("Claims ", clms, " are requested")

	if prmpts[prmptCons] {
		// 強制同意。
		log.Debug("Consenting that "+t.id+" gets ", clms, " is forced")

		if prmpts[prmptNone] {
			return redirectError(w, r, errConsReq, "cannot consent without UI")
		}
		return consent(sys, w, r, sess)
	} else if sess.hasNotConsented(sess.account(), t.id, clms) {
		// 同意が必要。
		log.Debug("Consenting that "+t.id+" gets ", clms, " is required")

		if prmpts[prmptNone] {
			return redirectError(w, r, errConsReq, "cannot consent without UI")
		}
		return consent(sys, w, r, sess)
	} else {
		// 同意済み。
		log.Debug("Consent that "+t.id+" gets ", clms, " is exist")
		return publishCode(sys, w, r, sess)
	}
}

func unauthenticated(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	// ユーザー認証前。

	t := r.ta()
	clms := r.claims()
	prmpts := r.prompts()

	if len(clms) == 0 {
		// 必要な同意は無い。
		log.Debug("No consent is required")
		if prmpts[prmptNone] {
			// UI 無しでユーザー認証だけする必要あり。
			log.Debug("Authenticate user without UI")
			return authenticateWithoutUi(sys, w, r, sess)
		} else {
			// ユーザー認証だけする必要あり。
			log.Debug("Authenticate user")
			return authenticate(sys, w, r, sess)
		}
	}

	// どのクレームが要求されてるか分かった。
	log.Debug("Claims ", clms, " are requested")

	if prmpts[prmptNone] {
		return redirectError(w, r, errConsReq, "cannot consent without UI")
	} else if prmpts[prmptCons] {
		// アカウント認証と強制同意。
		log.Debug("Consenting that "+t.id+" gets ", clms, " is forced")
		return authenticateAndConsent(sys, w, r, sess)
	} else {
		// 同意が必要。
		log.Debug("Consenting that "+t.id+" gets ", clms, " is required")
		return authenticateAndConsent(sys, w, r, sess)
	}
}

// ユーザー認証と同意処理。
func authenticateAndConsent(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	accName, passwd := r.authenticationData()
	if accName == "" || passwd == "" {
		// 認証情報入力前だった。
		log.Debug("Authentication data was not found")
		return redirectAuthenticationAndConsentUi(sys, w, r, sess)
	}

	// 認証情報を入力した後だった。
	log.Debug("Authentication data was found")

	acc, err := sys.accCont.getByName(accName)
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	} else if acc == nil {
		return redirectError(w, r, errAccDeny, "user "+accName+" is not exist")
	}
	if passwd != acc.Passwd {
		return redirectError(w, r, errAccDeny, "invalid password")
	}

	// 認証成功
	log.Debug("User " + accName + " (" + acc.Id + ") is authenticated")

	sess.setAccount(acc)
	return consent(sys, w, r, sess)
}

// ユーザー認証処理。
func authenticate(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	accName, passwd := r.authenticationData()
	if accName == "" || passwd == "" {
		// 認証情報入力前だった。
		log.Debug("Authentication data was not found")
		return redirectAuthenticationUi(sys, w, r, sess)
	}

	// 認証情報を入力した後だった。
	log.Debug("Authentication data was found")

	acc, err := sys.accCont.getByName(accName)
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	} else if acc == nil {
		return redirectError(w, r, errAccDeny, "user "+accName+" is not exist")
	}
	if passwd != acc.Passwd {
		return redirectError(w, r, errAccDeny, "invalid password")
	}

	// 認証成功
	log.Debug("User " + accName + " (" + acc.Id + ") is authenticated")

	sess.setAccount(acc)
	return publishCode(sys, w, r, sess)
}

// UI 無しユーザー認証処理。
func authenticateWithoutUi(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	accName, passwd := r.authenticationData()
	if accName == "" || passwd == "" {
		// 認証情報入力前だった。
		log.Debug("Authentication data was not found")
		return redirectError(w, r, errLoginReq, "cannot authenticate user without UI")
	}

	// 認証情報を入力した後だった。
	log.Debug("Authentication data was found")

	acc, err := sys.accCont.getByName(accName)
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	} else if acc == nil {
		return redirectError(w, r, errAccDeny, "user "+accName+" is not exist")
	}
	if passwd != acc.Passwd {
		return redirectError(w, r, errAccDeny, "invalid password")
	}

	// 認証成功
	log.Debug("User " + accName + " (" + acc.Id + ") is authenticated")

	sess.setAccount(acc)
	return publishCode(sys, w, r, sess)
}

// 同意処理。
func consent(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	consCod := r.consentCode()
	if consCod == "" {
		// アカウント選択 UI 表示前だった。
		log.Debug("Consent starts")
		return redirectConsentUi(sys, w, r, sess)
	}

	// 同意チケットがあった。
	log.Debug("Consent code " + mosaic(consCod) + " was declared")

	if consCod != sess.consentCode() {
		return redirectError(w, r, errAccDeny, "invalid account consent code")
	} else {
		// 同意 UI 表示後だった。
		log.Debug("Consent request returned")

		sess.consent(sess.account(), sess.accountName(), r.ta().id, r.claims())
		return publishCode(sys, w, r, sess)
	}
}

// 認可コード発行。
func publishCode(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {

	cod, err := sys.codCont.new(sess.account(), r.ta().id, r.redirectUri().String())
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	// 認可コードを発行した。
	log.Debug("Code " + mosaic(cod.Id) + " was published")

	if err := sys.sessCont.put(sess); err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	q := r.redirectUri().Query()
	q.Set(formCod, cod.Id)
	if stat := r.state(); stat != "" {
		q.Set(formStat, stat)
	}
	r.redirectUri().RawQuery = q.Encode()
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r.raw(), r.redirectUri().String(), http.StatusFound)
	return nil
}

// アカウント選択 UI にリダイレクトする。
func redirectAccountSelectionUi(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	selCod, err := util.SecureRandomString(sys.selCodLen)
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	// 選択コードを発行した。
	log.Debug("Selection code " + mosaic(selCod) + " was generated")

	sess.setSelectionCode(selCod)
	if err := sys.sessCont.put(sess); err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	q := r.raw().URL.Query()
	q.Set(formSelCod, selCod)
	// TODO 補助情報をつける。
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r.raw(), sys.uiUri+"?"+q.Encode(), http.StatusFound)
	return nil
}

// ユーザー認証と同意に UI にリダイレクトする。
func redirectAuthenticationAndConsentUi(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	consCod, err := util.SecureRandomString(sys.consCodLen)
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	// 選択コードを発行した。
	log.Debug("Consent code " + mosaic(consCod) + " was generated")

	sess.setConsentCode(consCod)
	if err := sys.sessCont.put(sess); err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	q := r.raw().URL.Query()
	q.Set(formConsCod, consCod)
	// TODO 補助情報をつける。
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r.raw(), sys.uiUri+"?"+q.Encode(), http.StatusFound)
	return nil
}

// ユーザー認証にリダイレクトする。
func redirectAuthenticationUi(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	if err := sys.sessCont.put(sess); err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	q := r.raw().URL.Query()
	// TODO 補助情報をつける。
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r.raw(), sys.uiUri+"?"+q.Encode(), http.StatusFound)
	return nil
}

// 同意にリダイレクトする。
func redirectConsentUi(sys *system, w http.ResponseWriter, r *authenticationRequest, sess *session) error {
	consCod, err := util.SecureRandomString(sys.consCodLen)
	if err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	// 選択コードを発行した。
	log.Debug("Consent code " + mosaic(consCod) + " was generated")

	sess.setConsentCode(consCod)
	if err := sys.sessCont.put(sess); err != nil {
		return redirectServerError(w, r, erro.Wrap(err))
	}

	log.Debug("Session " + mosaic(sess.id()) + " was registered")

	q := r.raw().URL.Query()
	q.Set(formConsCod, consCod)
	// TODO 補助情報をつける。
	http.SetCookie(w, &http.Cookie{
		Name:     cookSess,
		Value:    sess.id(),
		Expires:  sess.expirationDate(),
		Secure:   sys.secCook,
		HttpOnly: true})
	http.Redirect(w, r.raw(), sys.uiUri+"?"+q.Encode(), http.StatusFound)
	return nil
}
