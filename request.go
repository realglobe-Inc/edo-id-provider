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
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"strings"
)

// ブラウザからのリクエスト。
type browserRequest struct {
	sess string
}

func newBrowserRequest(r *http.Request) *browserRequest {
	var sess string
	if cook, err := r.Cookie(cookSess); err != nil {
		if err != http.ErrNoCookie {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
		}
	} else {
		sess = cook.Value
	}
	return &browserRequest{sess: sess}
}

func (this *browserRequest) session() string {
	return this.sess
}

// スペース区切りのフォーム値を集合にして返す。
func formValueSet(r *http.Request, key string) map[string]bool {
	s := r.FormValue(key)
	if s == "" {
		return nil
	}
	return strset.FromSlice(strings.Split(s, " "))
}

// フォーム値用にスペース区切りにして返す。
func valueSetToForm(m map[string]bool) string {
	buff := ""
	for v, ok := range m {
		if !ok || v == "" {
			continue
		}

		if len(buff) > 0 {
			buff += " "
		}
		buff += v
	}
	return buff
}

// スペース区切りのフォーム値を配列にして返す。
func formValues(r *http.Request, key string) []string {
	s := r.FormValue(key)
	if s == "" {
		return nil
	}
	return strings.Split(s, " ")
}

// フォーム値用にスペース区切りにして返す。
func valuesToForm(s []string) string {
	buff := ""
	for _, v := range s {
		if v == "" {
			continue
		}

		if len(buff) > 0 {
			buff += " "
		}
		buff += v
	}
	return buff
}
