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

// client_assertion に対応するデータ型。
package assertion

import (
	"encoding/json"
	"time"

	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
)

type Assertion struct {
	base *jwt.Jwt

	id  string
	iss string
	sub string
	aud map[string]bool
	exp time.Time
}

func Parse(raw []byte) (*Assertion, error) {
	base, err := jwt.Parse(raw)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	var buff struct {
		Iss string
		Sub string
		Aud audience.Audience
		Jti string
		Exp int64
	}
	if err := json.Unmarshal(base.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Iss == "" {
		return nil, erro.New("no issuer")
	} else if buff.Sub == "" {
		return nil, erro.New("no subject")
	} else if len(buff.Aud) == 0 {
		return nil, erro.New("no audience")
	} else if buff.Jti == "" {
		return nil, erro.New("no JWT ID")
	} else if buff.Exp == 0 {
		return nil, erro.New("no expiration date")
	}
	return &Assertion{
		base: base,
		id:   buff.Jti,
		iss:  buff.Iss,
		sub:  buff.Sub,
		aud:  buff.Aud,
		exp:  time.Unix(buff.Exp, 0),
	}, nil
}

func (this *Assertion) Id() string {
	return this.id
}

func (this *Assertion) Issuer() string {
	return this.iss
}

func (this *Assertion) Expires() time.Time {
	return this.exp
}

func (this *Assertion) Verify(taId string, taKeys []jwk.Key, aud string) error {
	if alg, _ := this.base.Header(tagAlg).(string); alg == tagNone {
		return erro.New("invalid algorithm " + alg)
	} else if this.iss != taId {
		return erro.New("invalid issuer")
	} else if this.sub != taId {
		return erro.New("invalid subject")
	} else if !this.aud[aud] {
		return erro.New("invalid audience")
	} else if time.Now().After(this.exp) {
		return erro.New("expired")
	}
	return this.base.Verify(taKeys)
}
