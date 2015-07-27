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

package auth

import (
	"net/http"

	"github.com/realglobe-Inc/go-lib/erro"
)

type loginRequest struct {
	tic      string
	acntName string
	pass     passInfo
	lang     string
}

func parseLoginRequest(r *http.Request) (*loginRequest, error) {
	tic := r.FormValue(tagTicket)
	if tic == "" {
		return nil, erro.New("no ticket")
	}
	acntName := r.FormValue(tagUsername)
	if acntName == "" {
		return nil, erro.New("no account name")
	}
	pass, err := parsePassInfo(r)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	for k, vs := range r.Form {
		if len(vs) != 1 {
			return nil, erro.New(k + " overlaps")
		}
	}

	return &loginRequest{
		tic:      tic,
		acntName: acntName,
		pass:     pass,
		lang:     r.FormValue(tagLocale),
	}, nil
}

func (this *loginRequest) ticket() string {
	return this.tic
}

func (this *loginRequest) accountName() string {
	return this.acntName
}

func (this *loginRequest) passInfo() passInfo {
	return this.pass
}

func (this *loginRequest) language() string {
	return this.lang
}
