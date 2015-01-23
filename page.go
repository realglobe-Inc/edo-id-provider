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
	grntTypeCod = "code"
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

const (
	errInvReq = iota
	errAccDeny
	errUnsuppRespType
	errInvScop
	errServErr

	errInteractReq
	errLoginReq
	errAccSelReq
	errConsReq
	errReqNotSupp
	errReqUriNotSupp
	errRegNotSupp

	errUnsuppGrntType
	errInvGrnt
	errInvTa

	errInvTok
)

var errCods []string = []string{
	errInvReq:         "invalid_request",
	errAccDeny:        "access_dnied",
	errUnsuppRespType: "unsupported_response_type",
	errInvScop:        "invalid_scope",
	errServErr:        "server_error",

	errInteractReq:   "interaction_required",
	errLoginReq:      "login_required",
	errAccSelReq:     "account_selection_required",
	errConsReq:       "consent_required",
	errReqNotSupp:    "request_not_supported",
	errReqUriNotSupp: "request_uri_not_supported",
	errRegNotSupp:    "registration_not_supported",

	errUnsuppGrntType: "unsupported_grant_type",
	errInvGrnt:        "invalid_grant",
	errInvTa:          "invalid_client",

	// OpenID Connect の仕様ではサンプルとしてしか登場しない。
	errInvTok: "invalid_token",
}

// リダイレクトしてエラーを通知する。
func redirectError(w http.ResponseWriter, r *http.Request, sys *system, sess *session, rediUri *url.URL, errCod int, errDesc string) error {
	if sess != nil && sess.id() != "" {
		sess.abort()
		if err := sys.sessCont.put(sess); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
		}
		log.Debug("Session " + mosaic(sess.id()) + " was aborted")
	}

	q := rediUri.Query()
	q.Set(formErr, errCods[errCod])
	if errDesc != "" {
		q.Set(formErrDesc, errDesc)
	}
	rediUri.RawQuery = q.Encode()
	http.Redirect(w, r, rediUri.String(), http.StatusFound)
	return nil
}

// リダイレクトしてサーバーエラーを通知する。
func redirectServerError(w http.ResponseWriter, r *http.Request, sys *system, sess *session, rediUri *url.URL, err error) error {
	log.Err(erro.Unwrap(err))
	log.Debug(err)
	return redirectError(w, r, sys, sess, rediUri, errServErr, erro.Unwrap(err).Error())
}
