// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
