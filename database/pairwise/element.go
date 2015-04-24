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

package pairwise

import (
	"crypto/sha256"
	"github.com/realglobe-Inc/edo-lib/base64url"
)

// TA 固有のアカウント ID の情報。
type Element struct {
	// 真のアカウント ID。
	acnt string
	// TA の ID。
	ta string
	// TA 固有のアカウント ID。
	pwAcnt string
}

func New(acnt, ta, pwAcnt string) *Element {
	return &Element{
		acnt:   acnt,
		ta:     ta,
		pwAcnt: pwAcnt,
	}
}

// 真のアカウント ID を返す。
func (this *Element) Account() string {
	return this.acnt
}

// TA の ID を返す。
func (this *Element) Ta() string {
	return this.ta
}

// TA 固有のアカウント ID を返す。
func (this *Element) PairwiseAccount() string {
	return this.pwAcnt
}

// TA 固有のアカウントを計算する。
func Generate(acnt, ta string) *Element {
	h := sha256.New()
	h.Write([]byte(ta))
	h.Write([]byte{0})
	h.Write([]byte(acnt))
	pwAcnt := base64url.EncodeToString(h.Sum(nil))
	return New(acnt, ta, pwAcnt)
}
