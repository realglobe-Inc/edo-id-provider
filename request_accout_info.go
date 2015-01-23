package main

import (
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"strings"
)

const (
	headAuth = "Authorization"
)

// アカウント情報リクエスト。
type accountInfoRequest struct {
	r *http.Request

	tok string
}

func newAccountInfoRequest(r *http.Request) (*accountInfoRequest, error) {
	req := &accountInfoRequest{r: r}

	if authLine := r.Header.Get(headAuth); authLine != "" {
		tok, err := parseAccountInfoRequestToken(authLine)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		req.tok = tok
	}

	return req, nil
}

func (this *accountInfoRequest) token() string {
	return this.tok
}

func parseAccountInfoRequestToken(line string) (tok string, err error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return "", erro.New("lack of parts")
	}
	sc := parts[0]
	rem := parts[1]

	switch sc {
	case scBear:
		return rem, nil
	default:
		return "", erro.New("scheme " + sc + " is not supported")
	}
}
