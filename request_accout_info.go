package main

import (
	"net/http"
	"strings"
)

// アカウント情報リクエスト。
type accountInfoRequest struct {
	sc  string
	tok string
}

func newAccountInfoRequest(r *http.Request) *accountInfoRequest {
	sc, tok := parseAuthorizationToken(r.Header.Get(headAuth))
	return &accountInfoRequest{
		sc:  sc,
		tok: tok,
	}
}

func (this *accountInfoRequest) scheme() string {
	return this.sc
}

func (this *accountInfoRequest) token() string {
	return this.tok
}

func parseAuthorizationToken(line string) (sc, tok string) {
	parts := strings.SplitN(line, " ", 2)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return "", parts[0]
	default:
		return parts[0], parts[1]
	}
}
