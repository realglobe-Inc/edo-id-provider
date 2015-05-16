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

package request

import (
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"strings"
)

// リクエストの基本情報。
type Request struct {
	sess string
	src  string
}

const (
	tagX_forwarded_for = "X-Forwarded-For"
)

func Parse(r *http.Request, sessLabel string) *Request {
	var sess string
	if sessLabel != "" {
		if cook, err := r.Cookie(sessLabel); err != nil {
			if err != http.ErrNoCookie {
				log.Warn(erro.Wrap(err))
			}
		} else {
			sess = cook.Value
		}
	}

	var src string
	if forwarded := r.Header.Get(tagX_forwarded_for); forwarded == "" {
		src = r.RemoteAddr
	} else {
		parts := strings.SplitN(forwarded, ",", 2)
		src = parts[0]
	}

	return &Request{
		sess: sess,
		src:  src,
	}
}

// Cookie で宣言されているセッション ID を返す。
func (this *Request) Session() string {
	return this.sess
}

// 送信元の IP アドレスを返す。
func (this *Request) Source() string {
	return this.src
}

var SessionDisplayLength = 8

// <セッション ID の最初の数文字>@<IP アドレス>
func (this *Request) String() string {
	str := ""
	if this.sess != "" {
		str += cutoff(this.sess, SessionDisplayLength) + "@"
	}
	return str + this.src
}

func cutoff(str string, max int) string {
	if len(str) <= max {
		return str
	} else {
		return str[:max]
	}
}
