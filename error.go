// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/realglobe-Inc/go-lib/erro"
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
