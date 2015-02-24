package main

import (
	"encoding/json"
	"fmt"
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
		MaxAge:     3600,
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
