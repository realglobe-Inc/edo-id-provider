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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"net/http"
)

type loginRequest struct {
	tic      string
	acntName string
	psType   string
	pass     passInfo
	lang     string
}

func newLoginRequest(r *http.Request) *loginRequest {
	psType := r.FormValue(formPass_type)

	var pass passInfo
	switch psType {
	case account.AuthTypeStr43:
		pass = newPasswordOnly(r.FormValue(formPassword))
	}

	return &loginRequest{
		tic:      r.FormValue(formTicket),
		acntName: r.FormValue(formUsername),
		psType:   psType,
		pass:     pass,
		lang:     r.FormValue(formLocale),
	}
}

func (this *loginRequest) ticket() string {
	return this.tic
}

func (this *loginRequest) accountName() string {
	return this.acntName
}

func (this *loginRequest) passType() string {
	return this.psType
}

func (this *loginRequest) passInfo() passInfo {
	return this.pass
}

func (this *loginRequest) language() string {
	return this.lang
}
