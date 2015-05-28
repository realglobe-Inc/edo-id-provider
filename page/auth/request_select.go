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
)

type selectRequest struct {
	tic      string
	acntName string
	lang     string
}

func newSelectRequest(r *http.Request) *selectRequest {
	return &selectRequest{
		tic:      r.FormValue(tagTicket),
		acntName: r.FormValue(tagUsername),
		lang:     r.FormValue(tagLocale),
	}
}

func (this *selectRequest) ticket() string {
	return this.tic
}

func (this *selectRequest) accountName() (accName string) {
	return this.acntName
}

func (this *selectRequest) language() string {
	return this.lang
}
