package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/util"
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

	tok *accountInfoRequestToken
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

func (this *accountInfoRequest) token() *accountInfoRequestToken {
	return this.tok
}

// アカウント情報リクエストに添付されたアクセストークン。
type accountInfoRequestToken struct {
	// 認証スキーム。Bearer とか。
	sc string
	// アクセストークン。
	tokId string
	// 認証スキームが JWS のときに使う。
	jws util.Jws
}

func parseAccountInfoRequestToken(line string) (*accountInfoRequestToken, error) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil, erro.New("lack of parts")
	}
	sc := parts[0]
	rem := parts[1]

	switch sc {
	case scBear:
		return &accountInfoRequestToken{sc: sc, tokId: rem}, nil
	case scJws:
		jws, err := util.ParseJws(rem)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		tokId, _ := jws.Claim(clmTok).(string)
		return &accountInfoRequestToken{sc: sc, tokId: tokId, jws: jws}, nil
	default:
		return nil, erro.New("scheme " + sc + " is not supported")
	}
}

func (this *accountInfoRequestToken) tokenId() string {
	return this.tokId
}

func (this *accountInfoRequestToken) scheme() string {
	return this.sc
}

func (this *accountInfoRequestToken) verify(keys map[string]crypto.PublicKey) error {
	switch this.sc {
	case scBear:
		return nil
	case scJws:
		if alg, _ := this.jws.Header(jwtAlg).(string); alg == algNone {
			return erro.New("JWS algorithm " + algNone + " is not allowed")
		}
		return this.jws.Verify(keys)
	default:
		return erro.New("scheme " + this.sc + " is not supported")
	}
}
