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
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"net/http"
)

type consentRequest struct {
	tic        string
	allowScop  map[string]bool
	allowAttrs map[string]bool
	denyScop   map[string]bool
	denyAttrs  map[string]bool
	lang       string
}

func newConsentRequest(r *http.Request) *consentRequest {
	return &consentRequest{
		tic:        r.FormValue(tagTicket),
		allowScop:  request.FormValueSet(r.FormValue(tagAllowed_scope)),
		allowAttrs: request.FormValueSet(r.FormValue(tagAllowed_claims)),
		denyScop:   request.FormValueSet(r.FormValue(tagDenied_scope)),
		denyAttrs:  request.FormValueSet(r.FormValue(tagDenied_claims)),
		lang:       r.FormValue(tagLocale),
	}
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
