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

package token

import (
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

type request struct {
	grntType  string
	cod       string
	ta_       string
	rediUri   string
	taAssType string
	taAss     []byte
}

func parseRequest(r *http.Request) (*request, error) {
	var taAss []byte
	if strTaAss := r.FormValue(tagClient_assertion); strTaAss != "" {
		taAss = []byte(strTaAss)
	}

	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return nil, erro.New(k + " overlaps")
		}
	}

	return &request{
		grntType:  r.FormValue(tagGrant_type),
		cod:       r.FormValue(tagCode),
		ta_:       r.FormValue(tagClient_id),
		rediUri:   r.FormValue(tagRedirect_uri),
		taAssType: r.FormValue(tagClient_assertion_type),
		taAss:     taAss,
	}, nil
}

func (this *request) grantType() string {
	return this.grntType
}

func (this *request) code() string {
	return this.cod
}

func (this *request) ta() string {
	return this.ta_
}

func (this *request) redirectUri() string {
	return this.rediUri
}

func (this *request) taAssertionType() string {
	return this.taAssType
}

func (this *request) taAssertion() []byte {
	return this.taAss
}
