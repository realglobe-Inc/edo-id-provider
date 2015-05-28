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

package coopto

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

type request struct {
	grntType string
	cod      string
	clmReq   *session.ClaimRequest
	subClms  map[string]session.Claims
	ass      []byte
}

func parseRequest(r *http.Request) (*request, error) {
	if r.Header.Get(tagContent_type) != contTypeJson {
		return nil, erro.New("not JSON")
	}

	var buff struct {
		GrntType string                    `json:"grant_type"`
		Cod      string                    `json:"code"`
		ClmReq   *session.ClaimRequest     `json:"claims"`
		SubClms  map[string]session.Claims `json:"user_claims"`
		Ass      string                    `json:"client_assertion"`
	}
	if err := json.NewDecoder(r.Body).Decode(&buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.GrntType == "" {
		return nil, erro.New("no grant type")
	} else if buff.Cod == "" {
		return nil, erro.New("no code")
	}
	var ass []byte
	if buff.Ass != "" {
		ass = []byte(buff.Ass)
	}
	return &request{
		grntType: buff.GrntType,
		cod:      buff.Cod,
		clmReq:   buff.ClmReq,
		subClms:  buff.SubClms,
		ass:      ass,
	}, nil
}

func (this *request) grantType() string {
	return this.grntType
}

func (this *request) code() string {
	return this.cod
}

func (this *request) claims() *session.ClaimRequest {
	return this.clmReq
}

func (this *request) subClaims() map[string]session.Claims {
	return this.subClms
}

func (this *request) taAssertion() []byte {
	return this.ass
}
