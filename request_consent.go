package main

import (
	"net/http"
)

type consentRequest struct {
	*browserRequest

	tic     string
	accName string
	passwd  string

	scops     map[string]bool
	clms      map[string]bool
	denyScops map[string]bool
	denyClms  map[string]bool
}

func newConsentRequest(r *http.Request) *consentRequest {
	return &consentRequest{
		browserRequest: newBrowserRequest(r),
		tic:            r.FormValue(formConsTic),
		scops:          stripUnknownScopes(formValueSet(r, formConsScops)),
		clms:           formValueSet(r, formConsClms),
		denyScops:      stripUnknownScopes(formValueSet(r, formDenyScops)),
		denyClms:       formValueSet(r, formDenyClms),
	}
}

func (this *consentRequest) ticket() string {
	return this.tic
}

func (this *consentRequest) consentInfo() (scops, clms, denyScops, denyClms map[string]bool) {
	return this.scops, this.clms, this.denyScops, this.denyClms
}
