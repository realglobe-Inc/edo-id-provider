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
	"testing"
)

func TestRequestSample(t *testing.T) {
	// OpenID Connect Core 1.0 Section 3.1.2.1 より。
	sample, err := http.NewRequest("GET", "https://server.example.com/authorize?"+
		"response_type=code"+
		"&scope=openid%20profile%20email"+
		"&client_id=s6BhdRkqt3"+
		"&state=af0ifjsldkj"+
		"&redirect_uri=https%3A%2F%2Fclient.example.org%2Fcb", nil)
	if err != nil {
		t.Error(err)
		return
	}

	req, err := ParseRequest(sample)
	if err != nil {
		t.Error(err)
		return
	}

	if respTyp := req.ResponseType(); !reflect.DeepEqual(respTyp, map[string]bool{"code": true}) {
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
