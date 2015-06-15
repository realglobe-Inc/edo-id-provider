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
	"strconv"
	"time"
)

const (
	test_id       = "XAOiyqgngWGzZbgl6j1w6Zm3ytHeI-"
	test_ticId    = "-TRO_YRa1B"
	test_lang     = "ja-JP"
	test_ta       = "https://ta.example.org"
	test_stat     = "YJgUit_Wx5"
	test_nonc     = "Wjj1_YUOlR"
	test_disp     = "page"
	test_maxAge   = 24 * time.Hour
	test_rediUri  = "https://ta.example.org/callback"
	test_acntId   = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"
	test_acntName = "edo-id-provider-tester"
)

var (
	test_acnt = NewAccount(test_acntId, test_acntName)
	test_req  *Request
)

func init() {
	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		panic(err)
	}
	q := url.Values{}
	q.Add("scope", "openid email")
	q.Add("response_type", "code id_token")
	q.Add("client_id", test_ta)
	q.Add("redirect_uri", test_rediUri)
	q.Add("state", test_stat)
	q.Add("nonce", test_nonc)
	q.Add("display", test_disp)
	q.Add("prompt", "login consent")
	q.Add("max_age", strconv.FormatInt(int64(test_maxAge/time.Second), 10))
	q.Add("ui_locales", "ja-JP")
	//q.Add("id_token_hint", "")
	q.Add("claims", `{"id_token":{"pds":{"essential":true}}}`)
	//q.Add("request", "")
	//q.Add("request_uri", "")
	r.URL.RawQuery = q.Encode()

	test_req, err = ParseRequest(r)
	if err != nil {
		panic(err)
	}
}
