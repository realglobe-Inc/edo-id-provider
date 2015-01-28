package main

import (
	"fmt"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
)

const (
	errInvReq         = "invalid_request"
	errAccDeny        = "access_dnied"
	errUnsuppRespType = "unsupported_response_type"
	errInvScop        = "invalid_scope"
	errServErr        = "server_error"
	errInteractReq    = "interaction_required"
	errLoginReq       = "login_required"
	errAccSelReq      = "account_selection_required"
	errConsReq        = "consent_required"
	errReqNotSupp     = "request_not_supported"
	errReqUriNotSupp  = "request_uri_not_supported"
	errRegNotSupp     = "registration_not_supported"
	errUnsuppGrntType = "unsupported_grant_type"
	errInvGrnt        = "invalid_grant"
	errInvTa          = "invalid_client"
	// OpenID Connect の仕様ではサンプルとしてしか登場しない。
	errInvTok = "invalid_token"
)

type idpError struct {
	// error の値。
	errCod string
	// error_description の値。
	errDesc string

	// 直接返すときの HTTP ステータス。
	stat int

	// 原因になったエラー。
	cause error
}

// stat が 0 の場合、代わりに http.StatusInternalServerError が入る。
func newIdpError(errCod string, errDesc string, stat int, cause error) error {
	if stat <= 0 {
		stat = http.StatusInternalServerError
	}
	if cause == nil {
		cause = erro.New(nil)
	}
	return &idpError{
		errCod:  errCod,
		errDesc: errDesc,
		stat:    stat,
		cause:   cause,
	}
}

func (this *idpError) Error() string {
	buff := this.errDesc
	if this.cause != nil {
		buff += fmt.Sprintln()
		buff += "caused by: "
		buff += this.cause.Error()
	}
	return buff
}

func (this *idpError) errorCode() string {
	return this.errCod
}

func (this *idpError) errorDescription() string {
	return this.errDesc
}

func (this *idpError) status() int {
	return this.stat
}

func (this *idpError) Cause() error {
	return this.cause
}
