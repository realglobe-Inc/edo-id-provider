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
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type authRequest struct {
	// リクエスト元 TA の ID。
	Ta string `json:"client_id"`
	// リクエスト元 TA の名前。
	TaName string `json:"client_name"`
	// 結果通知リダイレクト先。
	RawRediUri string `json:"redirect_uri"`
	rediUri    *url.URL
	// 結果の形式。
	RespType strset.StringSet `json:"response_type"`

	Stat    string           `json:"state,omitempty"`
	Nonc    string           `json:"nonce,omitempty"`
	Prmpts  strset.StringSet `json:"prompt,omitempty"`
	Scops   strset.StringSet `json:"scope,omitempty"`
	rawClms string
	Clms    claimRequestPair `json:"claims,omitempty"`
	Disp    string           `json:"display,omitempty"`
	UiLocs  []string         `json:"ui_localse,omitempty"`

	// 0 指定を許すために文字列で保存する。
	maxAgeParsed bool
	RawMaxAge    string `json:"max_age,omitempty"`
	maxAge_      int

	reqParsed bool
	rawReq    string
}

type claimRequestPair struct {
	AccInf claimRequest `json:"userinfo,omitempty"`
	IdTok  claimRequest `json:"id_token,omitempty"`
}

// エラーは idpError。
func newAuthRequest(r *http.Request) (*authRequest, error) {
	// TODO request_uri パラメータのサポート。

	return &authRequest{
		Ta:         r.FormValue(formTaId),
		RawRediUri: r.FormValue(formRediUri),
		RespType:   formValueSet(r, formRespType),
		Stat:       r.FormValue(formStat),
		Nonc:       r.FormValue(formNonc),
		Prmpts:     formValueSet(r, formPrmpt),
		Scops:      stripUnknownScopes(formValueSet(r, formScop)),
		rawClms:    r.FormValue(formClms),
		Disp:       r.FormValue(formDisp),
		UiLocs:     formValues(r, formUiLocs),
		RawMaxAge:  r.FormValue(formMaxAge),
		rawReq:     r.FormValue(formReq),
	}, nil
}

// リクエスト元 TA の ID を返す。
func (this *authRequest) ta() string {
	return this.Ta
}

// リクエスト元 TA 名を返す。
func (this *authRequest) taName() string {
	return this.TaName
}

// リクエスト元 TA 名を設定する。
func (this *authRequest) setTaName(taName string) {
	this.TaName = taName
}

// 結果を通知するリダイレクト先を返す。
func (this *authRequest) rawRedirectUri() string {
	return this.RawRediUri
}

func (this *authRequest) parseRedirectUri() error {
	if this.rediUri != nil || this.RawRediUri == "" {
		return nil
	}

	var err error
	this.rediUri, err = url.Parse(this.RawRediUri)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *authRequest) redirectUri() *url.URL {
	if this.rediUri == nil {
		this.parseRedirectUri()
	}
	return this.rediUri
}

// 結果の形式を返す。
func (this *authRequest) responseType() map[string]bool {
	if this.RespType == nil {
		this.RespType = strset.New(nil)
	}
	return this.RespType
}

// state 値を返す。
func (this *authRequest) state() string {
	return this.Stat
}

// nonce 値を返す。
func (this *authRequest) nonce() string {
	return this.Nonc
}

// 要求されている prompt を返す。
func (this *authRequest) prompts() map[string]bool {
	if this.Prmpts == nil {
		this.Prmpts = strset.New(nil)
	}
	return this.Prmpts
}

// 要求されている scope を返す。
func (this *authRequest) scopes() map[string]bool {
	if this.Scops == nil {
		this.Scops = strset.New(nil)
	}
	return this.Scops
}

// 要求されているクレームを返す。
func (this *authRequest) rawClaims() string {
	return this.rawClms
}

func (this *authRequest) parseClaims() error {
	if this.Clms.AccInf != nil || this.Clms.IdTok != nil || this.rawClms == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(this.rawClms), &this.Clms); err != nil {
		return erro.Wrap(err)
	}
	if this.Clms.AccInf == nil {
		this.Clms.AccInf = claimRequest{}
	}
	if this.Clms.IdTok == nil {
		this.Clms.IdTok = claimRequest{}
	}
	return nil
}

func (this *authRequest) claims() (accInfClms, idTokClms claimRequest) {
	if this.Clms.AccInf != nil || this.Clms.IdTok != nil {
		this.parseClaims()
	}
	return this.Clms.AccInf, this.Clms.IdTok
}

// 要求されているクレームを名前だけ返す。
func (this *authRequest) claimNames() map[string]bool {
	m := map[string]bool{}
	for _, clms := range []claimRequest{this.Clms.AccInf, this.Clms.IdTok} {
		for clmName := range clms {
			m[clmName] = true
		}
	}
	return m
}

// 要求されている表示形式を返す。
func (this *authRequest) display() string {
	return this.Disp
}

// 要求されている表示言語を優先する順に返す。
func (this *authRequest) uiLocales() []string {
	return this.UiLocs
}

// 過去の認証の有効期間を返す。
func (this *authRequest) rawMaxAge() string {
	return this.RawMaxAge
}

func (this *authRequest) parseMaxAge() error {
	if this.maxAgeParsed {
		return nil
	}

	if this.RawMaxAge == "" {
		this.maxAge_ = -1
		this.maxAgeParsed = true
		return nil
	}

	var err error
	this.maxAge_, err = strconv.Atoi(this.RawMaxAge)
	if err != nil {
		return erro.Wrap(err)
	}
	this.maxAgeParsed = true
	return nil
}

func (this *authRequest) maxAge() int {
	if !this.maxAgeParsed {
		this.parseMaxAge()
	}
	return this.maxAge_
}

// まとめて解析。
func (this *authRequest) parse() error {
	if err := this.parseRedirectUri(); err != nil {
		return erro.Wrap(err)
	} else if err := this.parseClaims(); err != nil {
		return erro.Wrap(err)
	} else if err := this.parseMaxAge(); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

// request パラメータを返す。
func (this *authRequest) rawRequest() string {
	return this.rawReq
}

func (this *authRequest) parseRequest(veriKeys map[string]interface{}, decKeys map[string]interface{}) error {
	if this.reqParsed || this.rawReq == "" {
		return nil
	}

	jt, err := jwt.Parse(this.rawReq, veriKeys, decKeys)
	if err != nil {
		return erro.Wrap(err)
	}

	authReq, err := authRequestFromJwt(jt)
	if err != nil {
		return erro.Wrap(err)
	}

	if authReq.Ta != "" && authReq.Ta != this.Ta {
		return erro.New(formTaId + " differs")
	} else if authReq.RespType != nil && !reflect.DeepEqual(authReq.RespType, this.RespType) {
		return erro.New(formRespType + " differs")
	}

	if authReq.RawRediUri != "" {
		this.RawRediUri = authReq.RawRediUri
	}
	if authReq.Stat != "" {
		this.Stat = authReq.Stat
	}
	if authReq.Nonc != "" {
		this.Nonc = authReq.Nonc
	}
	if authReq.Prmpts != nil {
		this.Prmpts = authReq.Prmpts
	}
	if authReq.Scops != nil {
		this.Scops = authReq.Scops
	}
	if authReq.Clms.AccInf != nil {
		this.Clms.AccInf = authReq.Clms.AccInf
	}
	if authReq.Clms.IdTok != nil {
		this.Clms.IdTok = authReq.Clms.IdTok
	}
	if authReq.Disp != "" {
		this.Disp = authReq.Disp
	}
	if authReq.UiLocs != nil {
		this.UiLocs = authReq.UiLocs
	}
	if authReq.RawMaxAge != "" {
		this.RawMaxAge = authReq.RawMaxAge
	}

	this.reqParsed = true
	return nil
}

func authRequestFromJwt(jt *jwt.Jwt) (authReq *authRequest, err error) {
	for ; jt.Nesting(); jt = jt.Nested() {
	}

	authReq = &authRequest{}

	if raw := jt.Claim(formTaId); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formTaId + " is not string")
		}
		authReq.Ta = str
	}
	if raw := jt.Claim(formRediUri); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formRediUri + " is not string")
		}
		authReq.RawRediUri = str
	}
	if raw := jt.Claim(formRespType); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formRespType + " is not string")
		}
		authReq.RespType = strset.FromSlice(strings.Split(str, " "))
	}
	if raw := jt.Claim(formStat); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formStat + " is not string")
		}
		authReq.Stat = str
	}
	if raw := jt.Claim(formNonc); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formNonc + " is not string")
		}
		authReq.Nonc = str
	}
	if raw := jt.Claim(formPrmpt); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formPrmpt + " is not string")
		}
		authReq.Prmpts = strset.FromSlice(strings.Split(str, " "))
	}
	if raw := jt.Claim(formScop); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formScop + " is not string")
		}
		authReq.Scops = strset.FromSlice(strings.Split(str, " "))
	}
	if raw := jt.Claim(formClms); raw != nil {
		m, ok := raw.(map[string]interface{})
		if !ok {
			log.Err("AHO ", raw)
			return nil, erro.New(formClms + " is not map")
		}
		if mAccInf := m["userinfo"]; mAccInf != nil {
			if authReq.Clms.AccInf, err = claimRequestFromMap(mAccInf); err != nil {
				return nil, erro.Wrap(err)
			}
		}
		if mIdTok := m["id_token"]; mIdTok != nil {
			if authReq.Clms.IdTok, err = claimRequestFromMap(mIdTok); err != nil {
				return nil, erro.Wrap(err)
			}
		}
	}
	if raw := jt.Claim(formDisp); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formDisp + " is not string")
		}
		authReq.Disp = str
	}
	if raw := jt.Claim(formUiLocs); raw != nil {
		str, ok := raw.(string)
		if !ok {
			return nil, erro.New(formUiLocs + " is not string")
		}
		authReq.UiLocs = strings.Split(str, " ")
	}
	if raw := jt.Claim(formMaxAge); raw != nil {
		val, ok := raw.(float64)
		if !ok {
			return nil, erro.New(formMaxAge + " is not number")
		}
		authReq.RawMaxAge = strconv.Itoa(int(val))
	}
	return authReq, nil
}
