package main

import (
	"net/http"
)

const (
	formConsTic   = "ticket"
	formConsScops = "consented_scope"
	formConsClms  = "consented_claim"
	formDenyScops = "denied_scope"
	formDenyClms  = "denied_claim"
)

type consentRequest struct {
	browserRequest

	tic     string
	accName string
	passwd  string

	scops     map[string]bool
	clms      map[string]bool
	denyScops map[string]bool
	denyClms  map[string]bool
}

func newConsentRequest(r *http.Request) *consentRequest {
	return &consentRequest{browserRequest: browserRequest{r: r}}
}

func (this *consentRequest) ticket() string {
	if this.tic == "" {
		this.tic = this.r.FormValue(formConsTic)
	}
	return this.tic
}

func (this *consentRequest) consentInfo() (scops, clms, denyScops, denyClms map[string]bool) {
	if this.scops == nil && this.clms == nil && this.denyScops == nil && this.denyClms == nil {
		this.scops = formValueSet(this.r, formConsScops)
		this.clms = formValueSet(this.r, formConsClms)
		this.denyScops = formValueSet(this.r, formDenyScops)
		this.denyClms = formValueSet(this.r, formDenyClms)
	}
	return this.scops, this.clms, this.denyScops, this.denyClms
}
