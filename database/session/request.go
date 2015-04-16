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

package session

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	labelScope         = "scope"
	labelResponse_type = "response_type"
	labelClient_id     = "client_id"
	labelRedirect_uri  = "redirect_uri"
	labelState         = "state"
	labelNonce         = "nonce"
	labelDisplay       = "display"
	labelPrompt        = "prompt"
	labelMax_age       = "max_age"
	labelUi_locales    = "ui_locales"
	labelId_token_hint = "id_token_hint"
	labelClaims        = "claims"
	labelRequest       = "request"
	labelRequest_uri   = "request_uri"
)

// セッションに付属させる認証リクエスト。
type Request struct {
	// scope
	scop map[string]bool
	// response_type
	respTyp map[string]bool
	// client_id。
	ta string
	// redirect_uri
	rediUri *url.URL
	// state
	stat string
	// nonc
	nonc string
	// display
	disp string
	// prompt
	prmpt map[string]bool
	// max_age
	maxAge time.Duration
	// ui_locales
	langs []string
	// id_token_hint
	hint string
	// claims
	clms *Claim
	// request
	req string
	// request_uri
	reqUri *url.URL
}

// 途中で失敗したら error と共にそこまでの結果も返す。
func ParseRequest(r *http.Request) (*Request, error) {
	var err error

	req := &Request{}

	req.scop = toSet(splitBySpace(r.FormValue(labelScope)))
	req.respTyp = toSet(splitBySpace(r.FormValue(labelResponse_type)))
	req.ta = r.FormValue(labelClient_id)
	if req.rediUri, err = url.ParseRequestURI(r.FormValue(labelRedirect_uri)); err != nil {
		return req, erro.Wrap(err)
	}
	req.stat = r.FormValue(labelState)
	req.nonc = r.FormValue(labelNonce)
	req.disp = r.FormValue(labelDisplay)
	req.prmpt = toSet(splitBySpace(r.FormValue(labelPrompt)))
	if req.maxAge, err = parseMaxAge(r.FormValue(labelMax_age)); err != nil {
		return req, erro.Wrap(err)
	}
	req.langs = splitBySpace(r.FormValue(labelUi_locales))
	req.hint = r.FormValue(labelId_token_hint)
	if req.clms, err = parseClaims(r.FormValue(labelClaims)); err != nil {
		return req, erro.Wrap(err)
	}
	req.req = r.FormValue(labelRequest)
	if reqUri := r.FormValue(labelRequest_uri); reqUri != "" {
		if req.reqUri, err = url.ParseRequestURI(reqUri); err != nil {
			return req, erro.Wrap(err)
		}
	}

	return req, nil
}

// scope を返す。
func (this *Request) Scope() map[string]bool {
	return this.scop
}

// response_type を返す。
func (this *Request) ResponseType() map[string]bool {
	return this.respTyp
}

// client_id を返す。
func (this *Request) Ta() string {
	return this.ta
}

// redirect_uri を返す。
func (this *Request) RedirectUri() *url.URL {
	return this.rediUri
}

// state を返す。
func (this *Request) State() string {
	return this.stat
}

// nonc を返す。
func (this *Request) Nonce() string {
	return this.nonc
}

// display を返す。
func (this *Request) Display() string {
	return this.disp
}

// prompt を返す。
func (this *Request) Prompt() map[string]bool {
	return this.prmpt
}

// max_age を返す。
// 未設定なら負値。
func (this *Request) MaxAge() time.Duration {
	return this.maxAge
}

// ui_locales を返す。
func (this *Request) Languages() []string {
	return this.langs
}

// id_token_hint を返す。
func (this *Request) IdTokenHint() string {
	return this.hint
}

// claims を返す。
func (this *Request) Claims() *Claim {
	return this.clms
}

// request を返す。
func (this *Request) Request() string {
	return this.req
}

// request_uri を返す。
func (this *Request) RequestUri() *url.URL {
	return this.reqUri
}

func toSet(a []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range a {
		m[s] = true
	}
	return m
}

// 1 つも無いときは要素 0 で返す。
func splitBySpace(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, " ")
}

// 未指定なら負値を返す。
func parseMaxAge(s string) (time.Duration, error) {
	if s == "" {
		return -1, nil
	}
	sec, err := strconv.Atoi(s)
	if err != nil {
		return 0, erro.Wrap(err)
	}
	return time.Duration(sec) * time.Second, nil
}

func parseClaims(s string) (*Claim, error) {
	if s == "" {
		return nil, nil
	}
	var clms Claim
	if err := json.Unmarshal([]byte(s), &clms); err != nil {
		return nil, erro.Wrap(err)
	}
	return &clms, nil
}
