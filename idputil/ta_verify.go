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

// ID プロバイダ周りの諸々。
package idputil

import (
	"encoding/json"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

type TaAssertion struct {
	base *jwt.Jwt

	iss string
	sub string
	aud map[string]bool
	id  string
	exp time.Time
}

func ParseTaAssertion(ass []byte) (*TaAssertion, error) {
	base, err := jwt.Parse(ass)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	var buff struct {
		Iss string
		Sub string
		Aud audience.Audience
		Id  string `json:"jti"`
		Exp int64
	}
	if err := json.Unmarshal(base.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	}
	return &TaAssertion{
		base: base,
		iss:  buff.Iss,
		sub:  buff.Sub,
		aud:  buff.Aud,
		id:   buff.Id,
		exp:  time.Unix(buff.Exp, 0),
	}, nil
}

func (this *TaAssertion) Issuer() string {
	return this.iss
}

func (this *TaAssertion) Subject() string {
	return this.sub
}

func (this *TaAssertion) Audience() map[string]bool {
	return this.aud
}

func (this *TaAssertion) Id() string {
	return this.id
}

func (this *TaAssertion) Expires() time.Time {
	return this.exp
}

func (this *TaAssertion) Verify(keys []jwk.Key, taId, aud string) error {
	if alg, _ := this.base.Header(tagAlg).(string); alg == tagNone {
		return erro.New("invalid algorithm " + alg)
	} else if this.iss != taId {
		return erro.New("issuer is " + this.iss + " not " + taId)
	} else if this.sub != taId {
		return erro.New("subject is " + this.sub + " not " + taId)
	} else if !this.aud[aud] {
		return erro.New("audience ", this.aud, " has no "+aud)
	} else if this.id == "" {
		return erro.New("no ID")
	} else if time.Now().After(this.exp) {
		return erro.New("expired")
	}
	return this.base.Verify(keys)
}

// client_assertion を検証する。
func VerifyAssertion(ass []byte, taId string, taKeys []jwk.Key, aud string) (*jtidb.Element, error) {
	jt, err := jwt.Parse(ass)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if alg, _ := jt.Header(tagAlg).(string); alg == tagNone {
		return nil, erro.New("invalid algorithm " + alg)
	} else if err := jt.Verify(taKeys); err != nil {
		return nil, erro.Wrap(err)
	}

	var buff struct {
		Iss string
		Sub string
		Aud audience.Audience
		Id  string `json:"jti"`
		Exp int64
	}
	if err := json.Unmarshal(jt.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Iss != taId {
		return nil, erro.New("invalid issuer " + buff.Iss)
	} else if buff.Sub != taId {
		return nil, erro.New("invalid subject " + buff.Sub)
	} else if !buff.Aud[aud] {
		return nil, erro.New("invalid audience ", buff.Aud)
	} else if buff.Id == "" {
		return nil, erro.New("no ID")
	} else if exp := time.Unix(buff.Exp, 0); time.Now().After(exp) {
		return nil, erro.New("expired")
	} else {
		return jtidb.New(taId, buff.Id, exp), nil
	}
}
