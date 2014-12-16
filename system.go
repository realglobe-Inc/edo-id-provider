package main

import (
	"regexp"
	"strings"
)

// 便宜的に集めただけ。
type system struct {
	uiUri  string
	uiPath string

	taCont   taContainer
	accCont  accountContainer
	sessCont sessionContainer
	codCont  codeContainer
	tokCont  tokenContainer
}

func newSystem(uiUri string, uiPath string,
	taCont taContainer,
	accCont accountContainer,
	sessCont sessionContainer,
	codCont codeContainer,
	tokCont tokenContainer) *system {

	uiUri = strings.TrimRight(uiUri, "/")
	uiUri = regexp.MustCompile("/+").ReplaceAllString(uiUri, "/")
	if uiUri == "" {
		uiUri = "/html"
	}
	if uiUri[0] != '/' {
		uiUri = "/" + uiUri
	}
	log.Info("Use " + uiUri + " as UI uri")
	return &system{uiUri, uiPath, taCont, accCont, sessCont, codCont, tokCont}

}
