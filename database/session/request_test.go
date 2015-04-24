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
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const (
	test_ta      = "https://ta.example.org"
	test_stat    = "YJgUit_Wx5"
	test_nonc    = "Wjj1_YUOlR"
	test_disp    = "page"
	test_max_age = 24 * time.Hour
)

var (
	test_redi_uri, _ = url.Parse("https://ta.example.org/return")
)

func TestRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		t.Error(err)
		return
	}
	q := url.Values{}
	q.Add("scope", "openid email")
	q.Add("response_type", "code id_token")
	q.Add("client_id", test_ta)
	q.Add("redirect_uri", test_redi_uri.String())
	q.Add("state", test_stat)
	q.Add("nonce", test_nonc)
	q.Add("display", test_disp)
	q.Add("prompt", "login consent")
	q.Add("max_age", strconv.FormatInt(int64(test_max_age/time.Second), 10))
	q.Add("ui_locales", "ja-JP")
	//q.Add("id_token_hint", "")
	q.Add("claims", `{"id_token":{"pds":{"essential":true}}}`)
	//q.Add("request", "")
	//q.Add("request_uri", "")
	r.URL.RawQuery = q.Encode()

	if req, err := ParseRequest(r); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(req.Scope(), map[string]bool{"openid": true, "email": true}) {
		t.Error(req.Scope())
	} else if !reflect.DeepEqual(req.ResponseType(), map[string]bool{"code": true, "id_token": true}) {
		t.Error(req.ResponseType())
	} else if req.Ta() != test_ta {
		t.Error(req.Ta())
		t.Error(test_ta)
	} else if !reflect.DeepEqual(req.RedirectUri(), test_redi_uri) {
		t.Error(req.RedirectUri())
		t.Error(test_redi_uri)
	} else if req.State() != test_stat {
		t.Error(req.State())
		t.Error(test_stat)
	} else if req.Nonce() != test_nonc {
		t.Error(req.Nonce())
		t.Error(test_nonc)
	} else if req.Display() != test_disp {
		t.Error(req.Display())
		t.Error(test_disp)
	} else if !reflect.DeepEqual(req.Prompt(), map[string]bool{"login": true, "consent": true}) {
		t.Error(req.Prompt())
	} else if req.MaxAge() != test_max_age {
		t.Error(req.MaxAge())
		t.Error(test_max_age)
	} else if !reflect.DeepEqual(req.Languages(), []string{"ja-JP"}) {
		t.Error(req.Languages())
	} else if clms := req.Claims(); clms == nil {
		t.Error("no claims")
	} else if wIdTok := clms.WithIdToken(); wIdTok == nil {
		t.Error("no id_token claims")
	} else if wIdTok := clms.WithIdToken(); wIdTok == nil {
		t.Error("no id_token claims")
	} else if wIdTokPds := wIdTok["pds"]; wIdTokPds == nil {
		t.Error("no id_token.pds claim")
	} else if !wIdTokPds.Essential() {
		t.Error("id_token.pds is not essential")
	}
}

func TestRequestSample(t *testing.T) {
	// OpenID Connect Core 1.0 Section 3.1.2.1 より。
	if r, err := http.NewRequest("GET", "https://server.example.com/authorize?"+
		"response_type=code"+
		"&scope=openid%20profile%20email"+
		"&client_id=s6BhdRkqt3"+
		"&state=af0ifjsldkj"+
		"&redirect_uri=https%3A%2F%2Fclient.example.org%2Fcb", nil); err != nil {
		t.Error(err)
	} else if req, err := ParseRequest(r); err != nil {
		t.Error(err)
	} else if respTyp := req.ResponseType(); !reflect.DeepEqual(respTyp, map[string]bool{"code": true}) {
		t.Error(respTyp)
	} else if scop := req.Scope(); !reflect.DeepEqual(scop, map[string]bool{"openid": true, "profile": true, "email": true}) {
		t.Error(scop)
	} else if ta := req.Ta(); ta != "s6BhdRkqt3" {
		t.Error(ta)
	} else if stat := req.State(); stat != "af0ifjsldkj" {
		t.Error(stat)
	} else if rediUri := req.RedirectUri(); rediUri == nil {
		t.Error("no redirect uri")
	} else if rediUri2, err := url.ParseRequestURI("https://client.example.org/cb"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rediUri, rediUri2) {
		t.Error(rediUri)
		t.Error(rediUri2)
	}
}
