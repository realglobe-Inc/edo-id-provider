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

package coopfrom

import (
	"encoding/json"
	requtil "github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

type request struct {
	grntType  string
	respType  map[string]bool
	frTa      string
	toTa_     string
	tok       string
	scop      map[string]bool
	expIn     time.Duration
	acntTag   string
	acnts     map[string]string
	hashAlg   string
	relAcnts  map[string]string
	relIdps   []string
	taAssType string
	taAss     []byte

	ref []byte
}

func parseRequest(r *http.Request) (*request, error) {
	if r.Header.Get(tagContent_type) != contTypeJson {
		return nil, erro.New("not json")
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, erro.Wrap(err)
	} else if len(req.respType) == 0 {
		return nil, erro.New("no response type")
	} else if req.grntType == "" {
		return nil, erro.New("no grant type")
	}
	return &req, nil
}

func (this *request) grantType() string {
	return this.grntType
}

func (this *request) responseType() map[string]bool {
	return this.respType
}

func (this *request) fromTa() string {
	return this.frTa
}

func (this *request) toTa() string {
	return this.toTa_
}

func (this *request) accessToken() string {
	return this.tok
}

func (this *request) scope() map[string]bool {
	return this.scop
}

func (this *request) expiresIn() time.Duration {
	return this.expIn
}

func (this *request) accountTag() string {
	return this.acntTag
}

func (this *request) accounts() map[string]string {
	return this.acnts
}

func (this *request) hashAlgorithm() string {
	return this.hashAlg
}

func (this *request) relatedAccounts() map[string]string {
	return this.relAcnts
}

func (this *request) relatedIdProviders() []string {
	return this.relIdps
}

func (this *request) taAssertionType() string {
	return this.taAssType
}

func (this *request) taAssertion() []byte {
	return this.taAss
}

func (this *request) referral() []byte {
	return this.ref
}

func (this *request) UnmarshalJSON(data []byte) error {
	var buff struct {
		GrntType  string            `json:"grant_type"`
		RespType  string            `json:"response_type"`
		FrTa      string            `json:"from_client"`
		ToTa      string            `json:"to_client"`
		Tok       string            `json:"access_token"`
		Scop      string            `json:"scope"`
		ExpIn     int               `json:"expires_in"`
		AcntTag   string            `json:"user_tag"`
		Acnts     map[string]string `json:"users"`
		HashAlg   string            `json:"hash_alg"`
		RelAcnts  map[string]string `json:"related_users"`
		RelIdps   []string          `json:"related_issuers"`
		TaAssType string            `json:"client_assertion_type"`
		TaAss     string            `json:"client_assertion"`
		Ref       string            `json:"referral"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	var scop map[string]bool
	if buff.Scop != "" {
		scop = requtil.FormValueSet(buff.Scop)
	}
	this.grntType = buff.GrntType
	this.respType = requtil.FormValueSet(buff.RespType)
	this.frTa = buff.FrTa
	this.toTa_ = buff.ToTa
	this.tok = buff.Tok
	this.scop = scop
	this.expIn = time.Duration(buff.ExpIn) * time.Second
	this.acntTag = buff.AcntTag
	this.acnts = buff.Acnts
	this.hashAlg = buff.HashAlg
	this.relAcnts = buff.RelAcnts
	this.relIdps = buff.RelIdps
	this.taAssType = buff.TaAssType
	if buff.TaAss != "" {
		this.taAss = []byte(buff.TaAss)
	}
	if buff.Ref != "" {
		this.ref = []byte(buff.Ref)
	}
	return nil
}
