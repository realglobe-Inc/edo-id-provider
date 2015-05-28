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
