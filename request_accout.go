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
	"strings"
)

// アカウント情報リクエスト。
type accountRequest struct {
	scm string
	tok string
}

func newAccountRequest(r *http.Request) *accountRequest {
	scm, tok := parseAuthorizationToken(r.Header.Get(headAuthorization))
	return &accountRequest{
		scm: scm,
		tok: tok,
	}
}

func (this *accountRequest) scheme() string {
	return this.scm
}

func (this *accountRequest) token() string {
	return this.tok
}

func parseAuthorizationToken(line string) (scm, tok string) {
	parts := strings.SplitN(line, " ", 2)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return "", parts[0]
	default:
		return parts[0], parts[1]
	}
}
