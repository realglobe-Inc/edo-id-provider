package main

import (
	"net/http"
)

type tokenRequest struct {
	r *http.Request

	grntType  string
	cod       string
	taid      string
	rediUri   string
	taAssType string
	taAss     string
}

func newTokenRequest(r *http.Request) *tokenRequest {
	return &tokenRequest{r: r}
}

func (this *tokenRequest) grantType() string {
	if this.grntType == "" {
		this.grntType = this.r.FormValue(formGrntType)
	}
	return this.grntType
}

func (this *tokenRequest) code() string {
	if this.cod == "" {
		this.cod = this.r.FormValue(formCod)
	}
	return this.cod
}

func (this *tokenRequest) taId() string {
	if this.taid == "" {
		this.taid = this.r.FormValue(formTaId)
	}
	return this.taid
}

func (this *tokenRequest) redirectUri() string {
	if this.rediUri == "" {
		this.rediUri = this.r.FormValue(formRediUri)
	}
	return this.rediUri
}

func (this *tokenRequest) taAssertionType() string {
	if this.taAssType == "" {
		this.taAssType = this.r.FormValue(formTaAssType)
	}
	return this.taAssType
}

func (this *tokenRequest) taAssertion() string {
	if this.taAss == "" {
		this.taAss = this.r.FormValue(formTaAss)
	}
	return this.taAss
}
