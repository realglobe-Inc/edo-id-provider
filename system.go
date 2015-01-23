package main

import (
	"github.com/realglobe-Inc/edo/util"
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

	tokExpiDur time.Duration

	sessExpiDur time.Duration
}

func (this *system) newTicket() (string, error) {
	log.Warn("Incomplete implementation")
	return util.SecureRandomString(10)
}
