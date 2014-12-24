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
	// prompt
	prmpts map[string]bool
	// 要求されているクレーム。
	clms map[string]bool

	sessId  string
	resType string
	selCod  string
	consCod string
	accName string
	passwd  string
	stat    string
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
	if this.sessId == "" {
		if cook, err := this.r.Cookie(cookSess); err != nil {
			if err != http.ErrNoCookie {
				err = erro.Wrap(err)
				log.Err(erro.Unwrap(err))
				log.Debug(err)
			}
		} else {
			this.sessId = cook.Value
		}
	}
	return this.sessId
}

func (this *authenticationRequest) responseType() string {
	if this.resType == "" {
		this.resType = this.r.FormValue(formRespType)
	}
	return this.resType
}

func (this *authenticationRequest) selectionCode() string {
	if this.selCod == "" {
		this.selCod = this.r.FormValue(formSelCod)
	}
	return this.selCod
}

func (this *authenticationRequest) consentCode() string {
	if this.consCod == "" {
		this.consCod = this.r.FormValue(formConsCod)
	}
	return this.consCod
}

func (this *authenticationRequest) authenticationData() (accName, passwd string) {
	if this.accName == "" && this.passwd == "" {
		this.accName, this.passwd = this.r.FormValue(formAccId), this.r.FormValue(formPasswd)
	}
	return this.accName, this.passwd
}

func (this *authenticationRequest) state() string {
	if this.stat == "" {
		this.stat = this.r.FormValue(formStat)
	}
	return this.stat
}
