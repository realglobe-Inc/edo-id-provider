package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/realglobe-Inc/edo/util/secrand"
	"github.com/realglobe-Inc/go-lib-rg/erro"
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
	sigKey crypto.PrivateKey
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
