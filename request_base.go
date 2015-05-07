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
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"strings"
)

// リクエストの基本情報。
type baseRequest struct {
	sess string
	src  string
}

func newBaseRequest(r *http.Request) *baseRequest {
	var sess string
	if cook, err := r.Cookie(sessLabel); err != nil {
		if err != http.ErrNoCookie {
			log.Err(erro.Wrap(err))
		}
	} else {
		sess = cook.Value
	}

	var src string
	if forwarded := r.Header.Get(headX_forwarded_for); forwarded == "" {
		src = r.RemoteAddr
	} else {
		parts := strings.SplitN(forwarded, ",", 2)
		src = parts[0]
	}

	return &baseRequest{
		sess: sess,
		src:  src,
	}
}

// Cookie の Id-Provider で宣言されているセッション ID を返す。
func (this *baseRequest) session() string {
	return this.sess
}

// 送信元の IP を返す。
func (this *baseRequest) source() string {
	return this.src
}
