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
	"encoding/json"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

type coopFromRequest struct {
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
}

func parseCoopFromRequest(r *http.Request) (*coopFromRequest, error) {
	if r.Header.Get(tagContent_type) != contTypeJson {
		return nil, erro.New("not json")
	}
	var req coopFromRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, erro.Wrap(err)
	}
	return &req, nil
}

func (this *coopFromRequest) grantType() string {
	return this.grntType
}

func (this *coopFromRequest) responseType() map[string]bool {
	return this.respType
}

func (this *coopFromRequest) fromTa() string {
	return this.frTa
}

func (this *coopFromRequest) toTa() string {
	return this.toTa_
}

func (this *coopFromRequest) accessToken() string {
	return this.tok
}

func (this *coopFromRequest) scope() map[string]bool {
	return this.scop
}

func (this *coopFromRequest) expiresIn() time.Duration {
	return this.expIn
}

func (this *coopFromRequest) accountTag() string {
	return this.acntTag
}

func (this *coopFromRequest) accounts() map[string]string {
	return this.acnts
}

func (this *coopFromRequest) hashAlgorithm() string {
	return this.hashAlg
}

func (this *coopFromRequest) relatedAccounts() map[string]string {
	return this.relAcnts
}

func (this *coopFromRequest) relatedIdProviders() []string {
	return this.relIdps
}

func (this *coopFromRequest) taAssertionType() string {
	return this.taAssType
}

func (this *coopFromRequest) taAssertion() []byte {
	return this.taAss
}

func (this *coopFromRequest) UnmarshalJSON(data []byte) error {
	var buff struct {
		GrntType  string            `json:"grant_type"`
		RespType  string            `json:"response_type"`
		FrTa      string            `json:"from_client"`
		ToTa      string            `json:"to_client"`
		Tok       string            `json:"access_token"`
		Scop      strset.Set        `json:"scope"`
		ExpIn     int               `json:"expires_in"`
		AcntTag   string            `json:"user_tag"`
		Acnts     map[string]string `json:"users"`
		HashAlg   string            `json:"hash_alg"`
		RelAcnts  map[string]string `json:"related_users"`
		RelIdps   []string          `json:"related_issuers"`
		TaAssType string            `json:"client_assertion_type"`
		TaAss     string            `json:"client_assertion"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.grntType = buff.GrntType
	this.respType = request.FormValueSet(buff.RespType)
	this.frTa = buff.FrTa
	this.toTa_ = buff.ToTa
	this.tok = buff.Tok
	this.scop = buff.Scop
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
	return nil
}
