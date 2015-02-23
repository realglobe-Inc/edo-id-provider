package main

// アカウント情報リクエスト・レスポンス周りのテスト。

import (
	"encoding/json"
	logutil "github.com/realglobe-Inc/edo/util/log"
	"github.com/realglobe-Inc/edo/util/server"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"testing"
	"time"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

// GET と POST でのアカウント情報リクエストに対応するか。
// Bearer 認証に対応するか。
// JSON を application/json で返すか。
// sub クレームを含むか。
func TestAccountInfo(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////

	testTa2, rediUri, kid, sigKey, taServ, idpSys, shutCh, err := setupTestTaAndIdp(nil, []*account{testAcc}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer taServ.Close()
	defer idpSys.close()
	defer os.RemoveAll(idpSys.uiPath)
	defer func() { shutCh <- struct{}{} }()
	// TA にリダイレクトしたときのレスポンスを設定しておく。
	taServ.AddResponse(http.StatusOK, nil, []byte("success"))

	cookJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	cli := &http.Client{Jar: cookJar}

	tokRes, err := testFromRequestAuthToGetToken(idpSys, cli, map[string]string{
		"scope":         "openid email",
		"response_type": "code",
		"client_id":     testTa2.id(),
		"redirect_uri":  rediUri,
	}, map[string]string{
		"username": testAcc.name(),
	}, map[string]string{
		"username": testAcc.name(),
		"password": testAcc.password(),
	}, map[string]string{
		"consented_scope": "openid email",
	}, map[string]interface{}{
		"alg": "RS256",
		"kid": kid,
	}, map[string]interface{}{
		"iss": testTa2.id(),
		"sub": testTa2.id(),
		"aud": idpSys.selfId + "/token",
		"jti": strconv.FormatInt(time.Now().UnixNano(), 16),
		"exp": time.Now().Add(idpSys.idTokExpiDur).Unix(),
	}, map[string]string{
		"grant_type":            "authorization_code",
		"redirect_uri":          rediUri,
		"client_id":             testTa2.id(),
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}, kid, sigKey)
	if err != nil {
		t.Fatal(err)
	}

	tok, _ := tokRes["access_token"].(string)
	if tok == "" {
		t.Fatal("no token")
	}

	for _, meth := range []string{"GET", "POST"} {
		req, err := http.NewRequest(meth, idpSys.selfId+"/userinfo", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Connection", "close")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(resp.StatusCode, http.StatusOK)
		} else if resp.Header.Get("Content-Type") != "application/json" {
			t.Error(resp.Header.Get("Content-Type"), "application/json")
		}

		var res struct{ Sub, Email string }
		if data, err := ioutil.ReadAll(resp.Body); err != nil {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if err := json.Unmarshal(data, &res); err != nil {
			server.LogRequest(level.ERR, req, true)
			server.LogResponse(level.ERR, resp, true)
			t.Fatal(err)
		} else if res.Sub != testAcc.id() {
			t.Fatal(res.Sub, testAcc.id())
		} else if em, _ := testAcc.attribute("email").(string); res.Email != em {
			t.Fatal(res.Email, em)
		}
	}
}
