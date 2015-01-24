package main

import (
	"net/http"
)

type tokenRequest struct {
	grntType  string
	cod       string
	taid      string
	rediUri   string
	taAssType string
	taAss     string
}

func newTokenRequest(r *http.Request) *tokenRequest {
	return &tokenRequest{
		grntType:  r.FormValue(formGrntType),
		cod:       r.FormValue(formCod),
		taid:      r.FormValue(formTaId),
		rediUri:   r.FormValue(formRediUri),
		taAssType: r.FormValue(formTaAssType),
		taAss:     r.FormValue(formTaAss),
	}
}

func (this *tokenRequest) grantType() string {
	return this.grntType
}

func (this *tokenRequest) code() string {
	return this.cod
}

func (this *tokenRequest) taId() string {
	return this.taid
}

func (this *tokenRequest) redirectUri() string {
	return this.rediUri
}

func (this *tokenRequest) taAssertionType() string {
	return this.taAssType
}

func (this *tokenRequest) taAssertion() string {
	return this.taAss
}
