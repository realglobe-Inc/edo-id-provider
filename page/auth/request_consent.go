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

package auth

import (
	"net/http"

	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/go-lib/erro"
)

type consentRequest struct {
	tic        string
	allowScop  map[string]bool
	allowAttrs map[string]bool
	denyScop   map[string]bool
	denyAttrs  map[string]bool
	lang       string
}

func parseConsentRequest(r *http.Request) (*consentRequest, error) {
	tic := r.FormValue(tagTicket)
	if tic == "" {
		return nil, erro.New("no ticket")
	}
	for k, vs := range r.Form {
		if len(vs) != 1 {
			return nil, erro.New(k + " overlaps")
		}
	}

	return &consentRequest{
		tic:        r.FormValue(tagTicket),
		allowScop:  request.FormValueSet(r.FormValue(tagAllowed_scope)),
		allowAttrs: request.FormValueSet(r.FormValue(tagAllowed_claims)),
		denyScop:   request.FormValueSet(r.FormValue(tagDenied_scope)),
		denyAttrs:  request.FormValueSet(r.FormValue(tagDenied_claims)),
		lang:       r.FormValue(tagLocale),
	}, nil
}

func (this *consentRequest) ticket() string {
	return this.tic
}

func (this *consentRequest) allowedScope() map[string]bool {
	return this.allowScop
}

func (this *consentRequest) allowedAttributes() map[string]bool {
	return this.allowAttrs
}

func (this *consentRequest) deniedScope() map[string]bool {
	return this.denyScop
}

func (this *consentRequest) deniedAttributes() map[string]bool {
	return this.denyAttrs
}

func (this *consentRequest) language() string {
	return this.lang
}
