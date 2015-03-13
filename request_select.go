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

type selectRequest struct {
	*browserRequest

	tic     string
	accName string
	loc     string
}

func newSelectRequest(r *http.Request) *selectRequest {
	return &selectRequest{
		browserRequest: newBrowserRequest(r),
		tic:            r.FormValue(formSelTic),
		accName:        r.FormValue(formAccName),
		loc:            r.FormValue(formLoc),
	}
}

func (this *selectRequest) ticket() string {
	return this.tic
}

func (this *selectRequest) selectInfo() (accName string) {
	return this.accName
}

func (this *selectRequest) locale() string {
	return this.loc
}
