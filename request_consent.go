package main

import (
	"net/http"
)

type consentRequest struct {
	*browserRequest

	tic       string
	scops     map[string]bool
	clms      map[string]bool
	denyScops map[string]bool
	denyClms  map[string]bool
	loc       string
}

func newConsentRequest(r *http.Request) *consentRequest {
	return &consentRequest{
		browserRequest: newBrowserRequest(r),
		tic:            r.FormValue(formConsTic),
		scops:          stripUnknownScopes(formValueSet(r, formConsScops)),
		clms:           formValueSet(r, formConsClms),
		denyScops:      stripUnknownScopes(formValueSet(r, formDenyScops)),
		denyClms:       formValueSet(r, formDenyClms),
		loc:            r.FormValue(formLoc),
	}
}

func (this *consentRequest) ticket() string {
	return this.tic
}

func (this *consentRequest) consentInfo() (scops, clms, denyScops, denyClms map[string]bool) {
	if this.scops == nil {
		this.scops = map[string]bool{}
	}
	if this.clms == nil {
		this.clms = map[string]bool{}
	}
	if this.denyScops == nil {
		this.denyScops = map[string]bool{}
	}
	if this.denyClms == nil {
		this.denyClms = map[string]bool{}
	}
	return this.scops, this.clms, this.denyScops, this.denyClms
}

func (this *consentRequest) locale() string {
	return this.loc
}
