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
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

// 認証情報。
type passInfo interface {
	passType() string
	params() []interface{}
}

func parsePassInfo(r *http.Request) (passInfo, error) {
	passType := r.FormValue(tagPass_type)
	switch passType {
	case tagPassword:
		passwd := r.FormValue(tagPassword)
		if passwd == "" {
			return nil, erro.New("no password")
		}
		return newPasswordInfo(passwd), nil
	default:
		return nil, erro.New("unsupported pass type " + passType)
	}
}

type passwordInfo struct {
	passwd string
}

func newPasswordInfo(passwd string) *passwordInfo {
	return &passwordInfo{passwd}
}

func (this *passwordInfo) passType() string {
	return tagPassword
}

func (this *passwordInfo) params() []interface{} {
	return []interface{}{this.passwd}
}
