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

package jti

import (
	"time"
)

// JWT の ID。
type Element struct {
	// JWT の iss。
	iss string
	// JWT の jti。
	jti string
	// JWT の exp または有効期限。
	exp time.Time
}

func New(iss, jti string, exp time.Time) *Element {
	return &Element{
		iss: iss,
		jti: jti,
		exp: exp,
	}
}

// iss を返す。
func (this *Element) Issuer() string {
	return this.iss
}

// jti を返す。
func (this *Element) Id() string {
	return this.jti
}

// exp または有効期限を返す。
func (this *Element) ExpiresIn() time.Time {
	return this.exp
}
