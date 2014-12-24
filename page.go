package main

import (
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
)

const cookSess = "X-Edo-Idp-Session"

const (
	// OAuth と OpenID Connect で定義されているパラメータ。
	formScop     = "scope"
	formTaId     = "client_id"
	formPrmpt    = "prompt"
	formRediUri  = "redirect_uri"
	formRespType = "response_type"
	formStat     = "state"
	formCod      = "code"
	formErr      = "error"
	formErrDesc  = "error_description"

	// 独自。
	formAccId   = "username"
	formPasswd  = "password"
	formSelCod  = "account_selection_code"
	formConsCod = "consent_code"
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
}

// rediUri にリダイレクトしてエラーを通知する。
func redirectError(w http.ResponseWriter, r errorRedirectRequest, errCod int, errDesc string) error {
	q := r.redirectUri().Query()
	q.Set(formErr, errCods[errCod])
	if errDesc != "" {
		q.Set(formErrDesc, errDesc)
	}
	r.redirectUri().RawQuery = q.Encode()
	http.Redirect(w, r.raw(), r.redirectUri().String(), http.StatusFound)
	return nil
}

// rediUri にリダイレクトしてサーバーエラーを通知する。
func redirectServerError(w http.ResponseWriter, r errorRedirectRequest, err error) error {
	log.Err(erro.Unwrap(err))
	log.Debug(err)
	return redirectError(w, r, errServErr, erro.Unwrap(err).Error())
}
