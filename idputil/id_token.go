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

package idputil

import (
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
)

type IdTokenSystem interface {
	SetSubSystem
	KeyDb() keydb.Db
	SignAlgorithm() string
	SignKeyId() string
	SelfId() string
	JwtIdExpiresIn() time.Duration
}

// ta に渡す acnt の ID トークンをつくる。
func IdToken(sys IdTokenSystem, ta tadb.Element, acnt account.Element, attrNames map[string]bool, clms map[string]interface{}) (string, error) {
	if err := SetSub(sys, acnt, ta); err != nil {
		return "", erro.Wrap(err)
	}
	keys, err := sys.KeyDb().Get()
	if err != nil {
		return "", erro.Wrap(err)
	}

	now := time.Now()
	jt := jwt.New()
	jt.SetHeader(tagAlg, sys.SignAlgorithm())
	if sys.SignKeyId() != "" {
		jt.SetHeader(tagKid, sys.SignKeyId())
	}
	jt.SetClaim(tagIss, sys.SelfId())
	jt.SetClaim(tagSub, acnt.Attribute(tagSub))
	jt.SetClaim(tagAud, audience.New(ta.Id()))
	jt.SetClaim(tagExp, now.Add(sys.JwtIdExpiresIn()).Unix())
	jt.SetClaim(tagIat, now.Unix())
	for k := range attrNames {
		jt.SetClaim(k, acnt.Attribute(k))
	}
	for k, v := range clms {
		jt.SetClaim(k, v)
	}

	if err := jt.Sign(keys); err != nil {
		return "", erro.Wrap(err)
	}
	buff, err := jt.Encode()
	if err != nil {
		return "", erro.Wrap(err)
	}
	return string(buff), nil
}
