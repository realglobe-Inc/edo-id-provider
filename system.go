package main

import (
)

// 便宜的に集めただけ。
type system struct {

	taCont   taContainer
	accCont  accountContainer
	sessCont sessionContainer
	codCont  codeContainer
	tokCont  tokenContainer
}

func newSystem(taCont taContainer,
	accCont accountContainer,
	sessCont sessionContainer,
	codCont codeContainer,
	tokCont tokenContainer) *system {

	return &system{taCont, accCont, sessCont, codCont, tokCont}
}
