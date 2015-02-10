package main

import (
	"github.com/realglobe-Inc/edo/util/strset"
	"net/http"
	"net/url"
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

	Stat   string                   `json:"state,omitempty"`
	Nonc   string                   `json:"nonce,omitempty"`
	Prmpts strset.StringSet         `json:"prompt,omitempty"`
	Scops  strset.StringSet         `json:"scope,omitempty"`
	Clms   map[string]*claimRequest `json:"claims,omitempty"`
	Disp   string                   `json:"display,omitempty"`
}

type claimRequest struct {
	Ess  bool          `json:"essential,omitempty"`
	Val  interface{}   `json:"value,omitempty"`
	Vals []interface{} `json:"values,omitempty"`

	Loc string `json:"locale,omitempty"`
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
		Clms:       map[string]*claimRequest{},
		Disp:       r.FormValue(formDisp),
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
func (this *authRequest) redirectUri() *url.URL {
	if this.rediUri == nil {
		this.rediUri, _ = url.Parse(this.RawRediUri)
	}
	return this.rediUri
}

// 結果を通知するリダイレクト先を設定する。
func (this *authRequest) setRedirectUri(rediUri *url.URL) {
	this.rediUri = rediUri
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
func (this *authRequest) claims() map[string]*claimRequest {
	return this.Clms
}

// 要求されているクレームを名前だけ返す。
func (this *authRequest) claimNames() map[string]bool {
	m := map[string]bool{}
	for clm := range this.Clms {
		m[clm] = true
	}
	return m
}

// 要求されている表示形式を返す。
func (this *authRequest) display() string {
	return this.Disp
}
