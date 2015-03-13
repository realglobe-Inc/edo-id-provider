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
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/realglobe-Inc/edo-lib/secrand"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

// 便宜的に集めただけ。
type system struct {
	selfId string

	secCook    bool
	selCodLen  int
	consCodLen int

	uiUri  string
	uiPath string

	taCont   taContainer
	accCont  accountContainer
	consCont consentContainer
	sessCont sessionContainer
	codCont  codeContainer
	tokCont  tokenContainer

	codExpiDur   time.Duration
	tokExpiDur   time.Duration
	idTokExpiDur time.Duration

	sessExpiDur time.Duration

	sigAlg string
	sigKid string
	sigKey interface{}
}

func (sys *system) newTicket() (string, error) {
	log.Warn("Incomplete implementation")
	return secrand.String(10)
}

func (sys *system) close() error {
	if err := sys.taCont.close(); err != nil {
		return erro.Wrap(err)
	} else if err := sys.accCont.close(); err != nil {
		return erro.Wrap(err)
	} else if err := sys.consCont.close(); err != nil {
		return erro.Wrap(err)
	} else if err := sys.sessCont.close(); err != nil {
		return erro.Wrap(err)
	} else if err := sys.codCont.close(); err != nil {
		return erro.Wrap(err)
	} else if err := sys.tokCont.close(); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (sys *system) verifyKey() crypto.PublicKey {
	switch key := sys.sigKey.(type) {
	case *rsa.PrivateKey:
		return &key.PublicKey
	case *ecdsa.PrivateKey:
		return &key.PublicKey
	default:
		return nil
	}
}
