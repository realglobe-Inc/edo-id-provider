package main

import (
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
	"strings"
)

// スペース区切りのフォーム値を集合にして返す。
func formValueSet(r *http.Request, key string) map[string]bool {
	s := r.FormValue(key)
	set := map[string]bool{}
	for _, v := range strings.Split(s, " ") {
		set[v] = true
	}
	return set
}

// 指定されているアカウント名を読む。
func getAccountId(r *http.Request) string {
	return r.FormValue(formAccId)
}

// 申告されている要求元 TA の ID を読む。
func getTaId(r *http.Request) string {
	return r.FormValue(formTaId)
}

// リダイレクト先 URI を読む。
func getRedirectUri(r *http.Request) string {
	return r.FormValue(formRediUri)
}

type authenticationRequest struct {
	// 元になる HTTP リクエスト。
	r *http.Request

	// 要求元の TA。
	t *ta
	// 処理後のリダイレクト先。
	rediUri *url.URL
	// scope
	scops map[string]bool
	// prmpt
	prmpts map[string]bool
	// 要求されているクレーム。
	clms map[string]bool
}

func newAuthenticationRequest(r *http.Request, t *ta, rediUri *url.URL) *authenticationRequest {
	log.Warn("Claim specification is not yet supported")
	return &authenticationRequest{
		r:       r,
		t:       t,
		rediUri: rediUri,
		scops:   formValueSet(r, formScop),
		prmpts:  formValueSet(r, formPrmpt),
		clms:    map[string]bool{},
	}
}

func (this *authenticationRequest) raw() *http.Request {
	return this.r
}

func (this *authenticationRequest) ta() *ta {
	return this.t
}

func (this *authenticationRequest) redirectUri() *url.URL {
	return this.rediUri
}

func (this *authenticationRequest) scopes() map[string]bool {
	return this.scops
}

func (this *authenticationRequest) prompts() map[string]bool {
	return this.prmpts
}

func (this *authenticationRequest) claims() map[string]bool {
	return this.clms
}

func (this *authenticationRequest) sessionId() string {
	if cook, err := this.r.Cookie(cookSess); err != nil {
		if err != http.ErrNoCookie {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
		}
		return ""
	} else {
		return cook.Value
	}
}

func (this *authenticationRequest) responseType() string {
	return this.r.FormValue(formRespType)
}

func (this *authenticationRequest) selectionCode() string {
	return this.r.FormValue(formSelCod)
}

func (this *authenticationRequest) consentCode() string {
	return this.r.FormValue(formConsCod)
}

func (this *authenticationRequest) authenticationData() (username, passwd string) {
	return this.r.FormValue(formAccId), this.r.FormValue(formPasswd)
}

func (this *authenticationRequest) state() string {
	return this.r.FormValue(formStat)
}
