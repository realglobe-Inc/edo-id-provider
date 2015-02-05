package main

import (
	"fmt"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
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
