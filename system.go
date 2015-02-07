package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/util/secrand"
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

func (this *system) newTicket() (string, error) {
	log.Warn("Incomplete implementation")
	return secrand.String(10)
}
