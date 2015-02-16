package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/util/strset"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
	"strconv"
)

type authRequest struct {
	// リクエスト元 TA の ID。
	Ta string `json:"client_id"`
	// リクエスト元 TA の名前。
	TaName string `json:"client_name"`
	// 結果通知リダイレクト先。
	RawRediUri string `json:"redirect_uri"`
	rediUri    *url.URL
	// 結果の形式。
	RespType strset.StringSet `json:"response_type"`

	Stat    string           `json:"state,omitempty"`
	Nonc    string           `json:"nonce,omitempty"`
	Prmpts  strset.StringSet `json:"prompt,omitempty"`
	Scops   strset.StringSet `json:"scope,omitempty"`
	rawClms string
	Clms    claimRequest `json:"claims,omitempty"`
	Disp    string       `json:"display,omitempty"`
	UiLocs  []string     `json:"ui_localse,omitempty"`

	rawMaxAge_ string
	MaxAge     int `json:"max_age,omitempty"`
}

// エラーは idpError。
func newAuthRequest(r *http.Request) (*authRequest, error) {
	// TODO claims, request, request_uri パラメータのサポート。

	return &authRequest{
		Ta:         r.FormValue(formTaId),
		RawRediUri: r.FormValue(formRediUri),
		RespType:   formValueSet(r, formRespType),
		Stat:       r.FormValue(formStat),
		Nonc:       r.FormValue(formNonc),
		Prmpts:     formValueSet(r, formPrmpt),
		Scops:      stripUnknownScopes(formValueSet(r, formScop)),
		rawClms:    r.FormValue(formClms),
		Disp:       r.FormValue(formDisp),
		UiLocs:     formValues(r, formUiLocs),
		rawMaxAge_: r.FormValue(formMaxAge),
	}, nil
}

// リクエスト元 TA の ID を返す。
func (this *authRequest) ta() string {
	return this.Ta
}

// リクエスト元 TA 名を返す。
func (this *authRequest) taName() string {
	return this.TaName
}

// リクエスト元 TA 名を設定する。
func (this *authRequest) setTaName(taName string) {
	this.TaName = taName
}

// 結果を通知するリダイレクト先を返す。
func (this *authRequest) rawRedirectUri() string {
	return this.RawRediUri
}

func (this *authRequest) parseRedirectUri() error {
	var err error
	this.rediUri, err = url.Parse(this.RawRediUri)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *authRequest) redirectUri() *url.URL {
	if this.rediUri == nil {
		this.parseRedirectUri()
	}
	return this.rediUri
}

// 結果の形式を返す。
func (this *authRequest) responseType() map[string]bool {
	return this.RespType
}

// state 値を返す。
func (this *authRequest) state() string {
	return this.Stat
}

// nonce 値を返す。
func (this *authRequest) nonce() string {
	return this.Nonc
}

// 要求されている prompt を返す。
func (this *authRequest) prompts() map[string]bool {
	return this.Prmpts
}

// 要求されている scope を返す。
func (this *authRequest) scopes() map[string]bool {
	return this.Scops
}

// 要求されているクレームを返す。
func (this *authRequest) rawClaims() string {
	return this.rawClms
}

func (this *authRequest) parseClaims() error {
	if err := json.Unmarshal([]byte(this.rawClms), &this.Clms); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *authRequest) claims() (accInfClms, idTokClms map[string]*claimUnit) {
	if this.Clms.accInf != nil || this.Clms.idTok != nil {
		this.parseClaims()
	}
	return this.Clms.accInf, this.Clms.idTok
}

// 要求されているクレームを名前だけ返す。
func (this *authRequest) claimNames() map[string]bool {
	m := map[string]bool{}
	for _, clms := range []map[string]*claimUnit{this.Clms.accInf, this.Clms.idTok} {
		for clmName := range clms {
			m[clmName] = true
		}
	}
	return m
}

// 要求されている表示形式を返す。
func (this *authRequest) display() string {
	return this.Disp
}

// 要求されている表示言語を優先する順に返す。
func (this *authRequest) uiLocales() []string {
	return this.UiLocs
}

// 過去の認証の有効期間を返す。
func (this *authRequest) rawMaxAge() string {
	return this.rawMaxAge_
}

func (this *authRequest) parseMaxAge() error {
	var err error
	this.MaxAge, err = strconv.Atoi(this.rawMaxAge_)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *authRequest) maxAge() int {
	if this.MaxAge == 0 {
		this.parseMaxAge()
	}
	return this.MaxAge
}
