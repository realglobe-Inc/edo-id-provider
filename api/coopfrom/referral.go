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

package coopfrom

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

type referral struct {
	base *jwt.Jwt

	id_      string
	idp      string
	frTa     string
	toTa_    string
	aud      map[string]bool
	relAcnts map[string]string
	hAlg     string
	exp      time.Time
}

func parseReferral(raw []byte) (*referral, error) {
	base, err := jwt.Parse(raw)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	var buff struct {
		Iss      string
		Sub      string
		Aud      audience.Audience
		Exp      int64
		Jti      string
		ToTa     string            `json:"to_client"`
		RelAcnts map[string]string `json:"related_users"`
		HAlg     string            `json:"hash_alg"`
	}
	if err := json.Unmarshal(base.RawBody(), &buff); err != nil {
		return nil, erro.Wrap(err)
	} else if buff.Iss == "" {
		return nil, erro.New("no ID provider")
	} else if buff.Sub == "" {
		return nil, erro.New("no from-TA")
	} else if len(buff.Aud) == 0 {
		return nil, erro.New("no audience")
	} else if buff.Exp == 0 {
		return nil, erro.New("no expiration date")
	} else if buff.Jti == "" {
		return nil, erro.New("no JWT ID")
	} else if buff.ToTa == "" {
		return nil, erro.New("no to-TA")
	} else if len(buff.RelAcnts) == 0 {
		return nil, erro.New("no related accounts")
	} else if buff.HAlg == "" {
		return nil, erro.New("no hash algorithm")
	}

	return &referral{
		base:     base,
		id_:      buff.Jti,
		idp:      buff.Iss,
		frTa:     buff.Sub,
		toTa_:    buff.ToTa,
		aud:      buff.Aud,
		relAcnts: buff.RelAcnts,
		hAlg:     buff.HAlg,
		exp:      time.Unix(buff.Exp, 0),
	}, nil
}

func (this *referral) id() string {
	return this.id_
}

func (this *referral) idProvider() string {
	return this.idp
}

func (this *referral) fromTa() string {
	return this.frTa
}

func (this *referral) toTa() string {
	return this.toTa_
}

func (this *referral) audience() map[string]bool {
	return this.aud
}

func (this *referral) relatedAccounts() map[string]string {
	return this.relAcnts
}

func (this *referral) hashAlgorithm() string {
	return this.hAlg
}

func (this *referral) expires() time.Time {
	return this.exp
}

func (this *referral) verify(keys []jwk.Key, aud string) error {
	if this.toTa_ == this.frTa {
		return erro.New("to-TA is from-TA")
	} else if !this.aud[aud] {
		return erro.New("invalid audience")
	} else if time.Now().After(this.exp) {
		return erro.New("expired")
	}
	return this.base.Verify(keys)
}
