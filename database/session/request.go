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
	"github.com/realglobe-Inc/edo-lib/duration"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"reflect"
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

const (
	tagCty = "cty"
)

const (
	ctyJwt = "JWT"
)

// セッションに付属させる認証リクエスト。
type Request struct {
	// scope
	scop map[string]bool
	// response_type
	respType map[string]bool
	// client_id
	ta string
	// redirect_uri
	rediUri string
	// state
	stat string
	// nonce
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
	reqClm *Claim
	// request
	req []byte
	// request_uri
	reqUri string
}

// 途中で失敗したら error と共にそこまでの結果も返す。
func ParseRequest(r *http.Request) (*Request, error) {
	var err error

	req := &Request{}

	req.scop = stringsToSet(splitBySpace(r.FormValue(labelScope)))
	req.respType = stringsToSet(splitBySpace(r.FormValue(labelResponse_type)))
	req.ta = r.FormValue(labelClient_id)
	req.rediUri = r.FormValue(labelRedirect_uri)
	req.stat = r.FormValue(labelState)
	req.nonc = r.FormValue(labelNonce)
	req.disp = r.FormValue(labelDisplay)
	req.prmpt = stringsToSet(splitBySpace(r.FormValue(labelPrompt)))
	if req.maxAge, err = parseMaxAge(r.FormValue(labelMax_age)); err != nil {
		return req, erro.Wrap(err)
	}
	req.langs = splitBySpace(r.FormValue(labelUi_locales))
	req.hint = r.FormValue(labelId_token_hint)
	if req.reqClm, err = parseClaims(r.FormValue(labelClaims)); err != nil {
		return req, erro.Wrap(err)
	}
	if reqObj := r.FormValue(labelRequest); reqObj != "" {
		req.req = []byte(reqObj)
	}
	req.reqUri = r.FormValue(labelRequest_uri)

	return req, nil
}

// scope を返す。
func (this *Request) Scope() map[string]bool {
	return this.scop
}

// response_type を返す。
func (this *Request) ResponseType() map[string]bool {
	return this.respType
}

// client_id を返す。
func (this *Request) Ta() string {
	return this.ta
}

// redirect_uri を返す。
func (this *Request) RedirectUri() string {
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
	return this.reqClm
}

// request を返す。
func (this *Request) Request() []byte {
	return this.req
}

// request_uri を返す。
func (this *Request) RequestUri() string {
	return this.reqUri
}

// requst や request_uri から取得したリクエストオブジェクトから読み込む。
func (this *Request) ParseRequest(req []byte, selfKeys, taKeys []jwk.Key) (err error) {
	var jt *jwt.Jwt
	for raw := req; ; {
		// 復号と検証。

		jt, err = jwt.Parse(raw)
		if err != nil {
			return erro.Wrap(err)
		}

		if jt.IsEncrypted() {
			if err := jt.Decrypt(selfKeys); err != nil {
				return erro.Wrap(err)
			}
			cty, _ := jt.Header(tagCty).(string)
			if cty == ctyJwt {
				raw = jt.RawBody()
				continue
			}
		}

		if jt.IsSigned() {
			if err := jt.Verify(taKeys); err != nil {
				return erro.Wrap(err)
			}
		}

		break
	}

	var buff struct {
		Scop     string           `json:"scope"`
		RespType string           `json:"response_type"`
		Ta       string           `json:"client_id"`
		RediUri  string           `json:"redirect_uri"`
		Stat     string           `json:"state"`
		Nonc     string           `json:"nonce"`
		Disp     string           `json:"display"`
		Prmpt    string           `json:"prompt"`
		MaxAge   *json.RawMessage `json:"max_age"`
		Langs    string           `json:"ui_locales"`
		Hint     string           `json:"id_token_hint"`
		ReqClm   *Claim           `json:"claims"`
		Req      string           `json:"request"`
		ReqUri   string           `json:"request_uri"`
	}
	if err := json.Unmarshal(jt.RawBody(), &buff); err != nil {
		return erro.Wrap(err)
	}

	if buff.Req != "" {
		return erro.New(labelRequest + " in request object")
	} else if buff.ReqUri != "" {
		return erro.New(labelRequest_uri + " in request object")
	} else if buff.RespType != "" && !reflect.DeepEqual(stringsToSet(splitBySpace(buff.RespType)), this.respType) {
		return erro.New("not same response type")
	} else if buff.Ta != "" && buff.Ta != this.ta {
		return erro.New("not same TA")
	}

	if buff.Scop != "" {
		this.scop = stringsToSet(splitBySpace(buff.Scop))
	}
	if buff.RediUri != "" {
		this.rediUri = buff.RediUri
	}
	if buff.Stat != "" {
		this.stat = buff.Stat
	}
	if buff.Nonc != "" {
		this.nonc = buff.Nonc
	}
	if buff.Disp != "" {
		this.disp = buff.Disp
	}
	if buff.Prmpt != "" {
		this.prmpt = stringsToSet(splitBySpace(buff.Prmpt))
	}
	if buff.MaxAge != nil {
		this.maxAge, err = parseMaxAge(string(*buff.MaxAge))
		if err != nil {
			return erro.Wrap(err)
		}
	}
	if buff.Langs != "" {
		this.langs = splitBySpace(buff.Langs)
	}
	if buff.Hint != "" {
		this.hint = buff.Hint
	}
	if buff.ReqClm != nil {
		this.reqClm = buff.ReqClm
	}

	return nil
}

func stringsToSet(a []string) map[string]bool {
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
	var reqClm Claim
	if err := json.Unmarshal([]byte(s), &reqClm); err != nil {
		return nil, erro.Wrap(err)
	}
	return &reqClm, nil
}

func SplitBySpace(s string) []string          { return splitBySpace(s) }
func StringsToSet(a []string) map[string]bool { return stringsToSet(a) }

func (this *Request) MarshalJSON() (data []byte, err error) {
	return json.Marshal(map[string]interface{}{
		"scope":         strset.Set(this.scop),
		"response_type": strset.Set(this.respType),
		"client_id":     this.ta,
		"redirect_uri":  this.rediUri,
		"state":         this.stat,
		"nonce":         this.nonc,
		"display":       this.disp,
		"prompt":        strset.Set(this.prmpt),
		"max_age":       duration.Duration(this.maxAge),
		"ui_locales":    this.langs,
		"id_token_hint": string(this.hint),
		"claims":        this.reqClm,
		// request と request_uri は使い切り。
	})
}

func (this *Request) UnmarshalJSON(data []byte) error {
	var buff struct {
		Scop    strset.Set        `json:"scope"`
		RespTyp strset.Set        `json:"response_type"`
		Ta      string            `json:"client_id"`
		RediUri string            `json:"redirect_uri"`
		Stat    string            `json:"state"`
		Nonc    string            `json:"nonce"`
		Disp    string            `json:"display"`
		Prmpt   strset.Set        `json:"prompt"`
		MaxAge  duration.Duration `json:"max_age"`
		Langs   []string          `json:"ui_locales"`
		Hint    string            `json:"id_token_hint"`
		ReqClm  *Claim            `json:"claims"`
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		return erro.Wrap(err)
	}

	this.scop = buff.Scop
	this.respType = buff.RespTyp
	this.ta = buff.Ta
	this.rediUri = buff.RediUri
	this.stat = buff.Stat
	this.nonc = buff.Nonc
	this.disp = buff.Disp
	this.prmpt = buff.Prmpt
	this.maxAge = time.Duration(buff.MaxAge)
	this.langs = buff.Langs
	this.hint = buff.Hint
	this.reqClm = buff.ReqClm
	return nil
}
