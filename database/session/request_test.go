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
	"github.com/realglobe-Inc/edo-id-provider/claims"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "https://idp.example.org/auth", nil)
	if err != nil {
		t.Fatal(err)
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

	if req, err := ParseRequest(r); err != nil {
		t.Fatal(err)
	} else if scop := map[string]bool{"openid": true, "email": true}; !reflect.DeepEqual(req.Scope(), scop) {
		t.Error(req.Scope())
		t.Fatal(scop)
	} else if respType := map[string]bool{"code": true, "id_token": true}; !reflect.DeepEqual(req.ResponseType(), respType) {
		t.Error(req.ResponseType())
		t.Fatal(respType)
	} else if req.Ta() != test_ta {
		t.Error(req.Ta())
		t.Fatal(test_ta)
	} else if req.RedirectUri() != test_rediUri {
		t.Error(req.RedirectUri())
		t.Fatal(test_rediUri)
	} else if req.State() != test_stat {
		t.Error(req.State())
		t.Fatal(test_stat)
	} else if req.Nonce() != test_nonc {
		t.Error(req.Nonce())
		t.Fatal(test_nonc)
	} else if req.Display() != test_disp {
		t.Error(req.Display())
		t.Fatal(test_disp)
	} else if prmpt := map[string]bool{"login": true, "consent": true}; !reflect.DeepEqual(req.Prompt(), prmpt) {
		t.Error(req.Prompt())
		t.Fatal(prmpt)
	} else if req.MaxAge() != test_maxAge {
		t.Error(req.MaxAge())
		t.Fatal(test_maxAge)
	} else if langs := []string{"ja-JP"}; !reflect.DeepEqual(req.Languages(), langs) {
		t.Error(req.Languages())
		t.Fatal(langs)
	} else if reqClm := req.Claims(); reqClm == nil {
		t.Fatal("no claims")
	} else if reqClm.AccountEntries() != nil {
		t.Fatal(reqClm.AccountEntries())
	} else if idTokClm := reqClm.IdTokenEntries(); idTokClm == nil {
		t.Fatal("no id_token claims")
	} else if pdsClm := idTokClm["pds"]; pdsClm == nil {
		t.Fatal("no id_token.pds claim")
	} else if !pdsClm.Essential() {
		t.Fatal("id_token.pds is not essential")
	}
}

func TestRequestSample1(t *testing.T) {
	// OpenID Connect Core 1.0 Section 3.1.2.1 より。
	if r, err := http.NewRequest("GET", "https://server.example.com/authorize?"+
		"response_type=code"+
		"&scope=openid%20profile%20email"+
		"&client_id=s6BhdRkqt3"+
		"&state=af0ifjsldkj"+
		"&redirect_uri=https%3A%2F%2Fclient.example.org%2Fcb", nil); err != nil {
		t.Fatal(err)
	} else if req, err := ParseRequest(r); err != nil {
		t.Fatal(err)
	} else if respType := map[string]bool{"code": true}; !reflect.DeepEqual(req.ResponseType(), respType) {
		t.Error(req.ResponseType())
		t.Fatal(respType)
	} else if scop := map[string]bool{"openid": true, "profile": true, "email": true}; !reflect.DeepEqual(req.Scope(), scop) {
		t.Error(req.Scope())
		t.Fatal(scop)
	} else if ta := "s6BhdRkqt3"; req.Ta() != ta {
		t.Error(req.Ta())
		t.Fatal(ta)
	} else if stat := "af0ifjsldkj"; req.State() != stat {
		t.Error(req.State())
		t.Fatal(stat)
	} else if rediUri := "https://client.example.org/cb"; req.RedirectUri() != rediUri {
		t.Error(req.RedirectUri())
		t.Fatal(rediUri)
	}
}

func TestRequestSample2(t *testing.T) {
	// OpenID Connect Core 1.0 Section 6.1 より。
	key, err := jwk.FromMap(map[string]interface{}{
		"kty": "RSA",
		"kid": "k2bdc",
		"n":   "y9Lqv4fCp6Ei-u2-ZCKq83YvbFEk6JMs_pSj76eMkddWRuWX2aBKGHAtKlE5P7_vn__PCKZWePt3vGkB6ePgzAFu08NmKemwE5bQI0e6kIChtt_6KzT5OaaXDFI6qCLJmk51Cc4VYFaxgqevMncYrzaW_50mZ1yGSFIQzLYP8bijAHGVjdEFgZaZEN9lsn_GdWLaJpHrB3ROlS50E45wxrlg9xMncVb8qDPuXZarvghLL0HzOuYRadBJVoWZowDNTpKpk2RklZ7QaBO7XDv3uR7s_sf2g-bAjSYxYUGsqkNA9b3xVW53am_UZZ3tZbFTIh557JICWKHlWj5uzeJXaw",
		"e":   "AQAB",
	})
	if err != nil {
		t.Fatal(err)
	}
	var reqClm claims.Request
	if err := json.Unmarshal([]byte(`{
    "userinfo": {
        "given_name": {"essential": true},
        "nickname": null,
        "email": {"essential": true},
        "email_verified": {"essential": true},
        "picture": null
    },
    "id_token": {
        "gender": null,
        "birthdate": {"essential": true},
        "acr": {"values": ["urn:mace:incommon:iap:silver"]}
    }
}`), &reqClm); err != nil {
		t.Fatal(err)
	}

	if r, err := http.NewRequest("GET", "https://server.example.com/authorize?response_type=code%20id_token"+
		"&client_id=s6BhdRkqt3"+
		"&request=eyJhbGciOiJSUzI1NiIsImtpZCI6ImsyYmRjIn0.ew0KICJpc3MiOiAiczZCaGRSa3F0MyIsDQogImF1ZCI6ICJodHRwczovL3NlcnZlci5leGFtcGxlLmNvbSIsDQogInJlc3BvbnNlX3R5cGUiOiAiY29kZSBpZF90b2tlbiIsDQogImNsaWVudF9pZCI6ICJzNkJoZFJrcXQzIiwNCiAicmVkaXJlY3RfdXJpIjogImh0dHBzOi8vY2xpZW50LmV4YW1wbGUub3JnL2NiIiwNCiAic2NvcGUiOiAib3BlbmlkIiwNCiAic3RhdGUiOiAiYWYwaWZqc2xka2oiLA0KICJub25jZSI6ICJuLTBTNl9XekEyTWoiLA0KICJtYXhfYWdlIjogODY0MDAsDQogImNsYWltcyI6IA0KICB7DQogICAidXNlcmluZm8iOiANCiAgICB7DQogICAgICJnaXZlbl9uYW1lIjogeyJlc3NlbnRpYWwiOiB0cnVlfSwNCiAgICAgIm5pY2tuYW1lIjogbnVsbCwNCiAgICAgImVtYWlsIjogeyJlc3NlbnRpYWwiOiB0cnVlfSwNCiAgICAgImVtYWlsX3ZlcmlmaWVkIjogeyJlc3NlbnRpYWwiOiB0cnVlfSwNCiAgICAgInBpY3R1cmUiOiBudWxsDQogICAgfSwNCiAgICJpZF90b2tlbiI6IA0KICAgIHsNCiAgICAgImdlbmRlciI6IG51bGwsDQogICAgICJiaXJ0aGRhdGUiOiB7ImVzc2VudGlhbCI6IHRydWV9LA0KICAgICAiYWNyIjogeyJ2YWx1ZXMiOiBbInVybjptYWNlOmluY29tbW9uOmlhcDpzaWx2ZXIiXX0NCiAgICB9DQogIH0NCn0.nwwnNsk1-ZkbmnvsF6zTHm8CHERFMGQPhos-EJcaH4Hh-sMgk8ePrGhw_trPYs8KQxsn6R9Emo_wHwajyFKzuMXZFSZ3p6Mb8dkxtVyjoy2GIzvuJT_u7PkY2t8QU9hjBcHs68PkgjDVTrG1uRTx0GxFbuPbj96tVuj11pTnmFCUR6IEOXKYr7iGOCRB3btfJhM0_AKQUfqKnRlrRscc8Kol-cSLWoYE9l5QqholImzjT_cMnNIznW9E7CDyWXTsO70xnB4SkG6pXfLSjLLlxmPGiyon_-Te111V8uE83IlzCYIb_NMXvtTIVc1jpspnTSD7xMbpL-2QgwUsAlMGzw", nil); err != nil {
		t.Fatal(err)
	} else if req, err := ParseRequest(r); err != nil {
		t.Fatal(err)
	} else if req.Request() == nil {
		t.Fatal("cannot parse request object")
	} else if err := req.ParseRequest(req.Request(), nil, []jwk.Key{key}); err != nil {
		t.Fatal(err)
	} else if respType := map[string]bool{"code": true, "id_token": true}; !reflect.DeepEqual(req.ResponseType(), respType) {
		t.Error(req.ResponseType())
		t.Fatal(respType)
	} else if ta := "s6BhdRkqt3"; req.Ta() != ta {
		t.Error(req.Ta())
		t.Fatal(ta)
	} else if rediUri := "https://client.example.org/cb"; req.RedirectUri() != rediUri {
		t.Error(req.RedirectUri())
		t.Fatal(rediUri)
	} else if scop := map[string]bool{"openid": true}; !reflect.DeepEqual(req.Scope(), scop) {
		t.Error(req.Scope())
		t.Fatal(scop)
	} else if stat := "af0ifjsldkj"; req.State() != stat {
		t.Error(req.State())
		t.Fatal(stat)
	} else if nonc := "n-0S6_WzA2Mj"; req.Nonce() != nonc {
		t.Error(req.Nonce())
		t.Fatal(nonc)
	} else if maxAge := 86400 * time.Second; req.MaxAge() != maxAge {
		t.Error(req.MaxAge())
		t.Fatal(maxAge)
	} else if !reflect.DeepEqual(req.Claims(), &reqClm) {
		t.Error(req.Claims())
		t.Fatal(&reqClm)
	}
}
