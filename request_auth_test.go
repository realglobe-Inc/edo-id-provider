package main

import (
	"encoding/json"
	"fmt"
	"github.com/realglobe-Inc/edo-toolkit/util/jwt"
	logutil "github.com/realglobe-Inc/edo-toolkit/util/log"
	"github.com/realglobe-Inc/edo-toolkit/util/strset"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"reflect"
	"testing"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

func TestAuthRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////
	req := &authRequest{
		Ta:         "test_ta_id",
		TaName:     "test_ta_name",
		RawRediUri: "http://test.example.org/redirect_uri",
		RespType:   strset.FromSlice([]string{"code"}),
		Stat:       "test_state",
		Nonc:       "test_nonce",
		Prmpts:     strset.FromSlice([]string{"login", "consent"}),
		Scops:      strset.FromSlice([]string{"openid", "email"}),
		Disp:       "page",
		UiLocs:     []string{"ja"},
		RawMaxAge:  "3600",
	}
	req.Clms.AccInf = claimRequest{
		"name":     {"ja": {true, nil, nil}},
		"nickname": {"": {false, nil, nil}},
	}

	buff, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var req2 authRequest
	if err := json.Unmarshal(buff, &req2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(&req2, req) {
		t.Error(fmt.Sprintf("%#v", &req2))
		t.Error(fmt.Sprintf("%#v", req))
		t.Error(string(buff))
	}
}

func TestAuthRequestParseRequest(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	// OpenID Connect Core 1.0 6.1 より。
	req := &authRequest{
		Ta:       "s6BhdRkqt3",
		RespType: strset.FromSlice([]string{"code", "id_token"}),
		rawReq:   "eyJhbGciOiJSUzI1NiIsImtpZCI6ImsyYmRjIn0.ew0KICJpc3MiOiAiczZCaGRSa3F0MyIsDQogImF1ZCI6ICJodHRwczovL3NlcnZlci5leGFtcGxlLmNvbSIsDQogInJlc3BvbnNlX3R5cGUiOiAiY29kZSBpZF90b2tlbiIsDQogImNsaWVudF9pZCI6ICJzNkJoZFJrcXQzIiwNCiAicmVkaXJlY3RfdXJpIjogImh0dHBzOi8vY2xpZW50LmV4YW1wbGUub3JnL2NiIiwNCiAic2NvcGUiOiAib3BlbmlkIiwNCiAic3RhdGUiOiAiYWYwaWZqc2xka2oiLA0KICJub25jZSI6ICJuLTBTNl9XekEyTWoiLA0KICJtYXhfYWdlIjogODY0MDAsDQogImNsYWltcyI6IA0KICB7DQogICAidXNlcmluZm8iOiANCiAgICB7DQogICAgICJnaXZlbl9uYW1lIjogeyJlc3NlbnRpYWwiOiB0cnVlfSwNCiAgICAgIm5pY2tuYW1lIjogbnVsbCwNCiAgICAgImVtYWlsIjogeyJlc3NlbnRpYWwiOiB0cnVlfSwNCiAgICAgImVtYWlsX3ZlcmlmaWVkIjogeyJlc3NlbnRpYWwiOiB0cnVlfSwNCiAgICAgInBpY3R1cmUiOiBudWxsDQogICAgfSwNCiAgICJpZF90b2tlbiI6IA0KICAgIHsNCiAgICAgImdlbmRlciI6IG51bGwsDQogICAgICJiaXJ0aGRhdGUiOiB7ImVzc2VudGlhbCI6IHRydWV9LA0KICAgICAiYWNyIjogeyJ2YWx1ZXMiOiBbInVybjptYWNlOmluY29tbW9uOmlhcDpzaWx2ZXIiXX0NCiAgICB9DQogIH0NCn0.nwwnNsk1-ZkbmnvsF6zTHm8CHERFMGQPhos-EJcaH4Hh-sMgk8ePrGhw_trPYs8KQxsn6R9Emo_wHwajyFKzuMXZFSZ3p6Mb8dkxtVyjoy2GIzvuJT_u7PkY2t8QU9hjBcHs68PkgjDVTrG1uRTx0GxFbuPbj96tVuj11pTnmFCUR6IEOXKYr7iGOCRB3btfJhM0_AKQUfqKnRlrRscc8Kol-cSLWoYE9l5QqholImzjT_cMnNIznW9E7CDyWXTsO70xnB4SkG6pXfLSjLLlxmPGiyon_-Te111V8uE83IlzCYIb_NMXvtTIVc1jpspnTSD7xMbpL-2QgwUsAlMGzw",
	}
	key, err := jwt.KeyFromJwkMap(map[string]interface{}{
		"kty": "RSA",
		"kid": "k2bdc",
		"n":   "y9Lqv4fCp6Ei-u2-ZCKq83YvbFEk6JMs_pSj76eMkddWRuWX2aBKGHAtKlE5P7_vn__PCKZWePt3vGkB6ePgzAFu08NmKemwE5bQI0e6kIChtt_6KzT5OaaXDFI6qCLJmk51Cc4VYFaxgqevMncYrzaW_50mZ1yGSFIQzLYP8bijAHGVjdEFgZaZEN9lsn_GdWLaJpHrB3ROlS50E45wxrlg9xMncVb8qDPuXZarvghLL0HzOuYRadBJVoWZowDNTpKpk2RklZ7QaBO7XDv3uR7s_sf2g-bAjSYxYUGsqkNA9b3xVW53am_UZZ3tZbFTIh557JICWKHlWj5uzeJXaw",
		"e":   "AQAB",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := req.parseRequest(map[string]interface{}{"k2bdc": key}, nil); err != nil {
		t.Fatal(err)
	} else if req.state() != "af0ifjsldkj" {
		t.Error(fmt.Sprintf("%#v", req))
	}
}
