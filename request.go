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
	"github.com/realglobe-Inc/edo-id-provider/database/session"
)

// リクエスト解析用関数。

// "openid email" みたいな文字列を
// {"openid": true, "email": true} みたいな集合にする。
func formValueSet(s string) map[string]bool {
	return session.StringsToSet(session.SplitBySpace(s))
}

// {"openid": true, "email": true} みたいな集合を
// "openid email" みたいな文字列にする
func valueSetForm(m map[string]bool) string {
	buff := ""
	for v := range m {
		if len(buff) > 0 {
			buff += " "
		}
		buff += v
	}
	return buff
}

// "openid email" みたいな文字列を
// {"openid", "email"} みたいな配列にする。
var formValues = session.SplitBySpace

// {"openid", "email"} みたいな配列を
// "openid email" みたいな文字列にする
func valuesForm(a []string) string {
	buff := ""
	for _, v := range a {
		if len(buff) > 0 {
			buff += " "
		}
		buff += v
	}
	return buff
}
