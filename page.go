package main

import (
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
)

const cookSess = "X-Edo-Idp-Session"

const (
	// OAuth と OpenID Connect で定義されているパラメータ。
	formScop      = "scope"
	formTaId      = "client_id"
	formPrmpt     = "prompt"
	formRediUri   = "redirect_uri"
	formRespType  = "response_type"
	formStat      = "state"
	formNonc      = "nonce"
	formCod       = "code"
	formGrntType  = "grant_type"
	formTaAssType = "client_assertion_type"
	formTaAss     = "client_assertion"
	formTokId     = "access_token"
	formTokType   = "token_type"
	formExpi      = "expires_in"
	formRefTok    = "refresh_token"
	formIdTok     = "id_token"
	formErr       = "error"
	formErrDesc   = "error_description"

	// 独自。
	formAccName = "username"
	formPasswd  = "password"
)

const (
	scopOpId = "openid"
)

const (
	respTypeCod = "code"
)

const (
	prmptSelAcc = "select_account"
	prmptLogin  = "login"
	prmptCons   = "consent"
	prmptNone   = "none"
)

const (
	taAssTypeJwt = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

const (
	clmIss     = "iss"
	clmSub     = "sub"
	clmAud     = "aud"
	clmJti     = "jti"
	clmExp     = "exp"
	clmIat     = "iat"
	clmAuthTim = "auth_time"
	clmNonc    = "nonce"

	// プライベートクレーム。
	clmCod = "code"
	clmTok = "access_token"
)

const (
	tokTypeBear = "Bearer"
)

func redirectError(w http.ResponseWriter, r *http.Request, sys *system, sess *session, rediUri *url.URL, err error) error {
	if sess != nil && sess.id() != "" {
		// 認証経過を廃棄。
		sess.abort()
		if err := sys.sessCont.put(sess); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
		} else {
			log.Debug("Session " + mosaic(sess.id()) + " was aborted")
		}
	}

	q := rediUri.Query()
	switch e := erro.Unwrap(err).(type) {
	case *idpError:
		log.Err(e.errorDescription())
		log.Debug(e)
		q.Set(formErr, e.errorCode())
		q.Set(formErrDesc, e.errorDescription())
	default:
		log.Err(e)
		log.Debug(err)
		q.Set(formErr, errServErr)
		q.Set(formErrDesc, e.Error())
	}

	rediUri.RawQuery = q.Encode()
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	http.Redirect(w, r, rediUri.String(), http.StatusFound)
	return nil
}
