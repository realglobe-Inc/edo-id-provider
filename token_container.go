package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type tokenContainer interface {
	new(cod *code) (*token, error)
	get(tokId string) (*token, error)
}

type tokenContainerImpl struct {
	// 認可コードの文字数。
	idLen int
	// 自分の IdP としての ID。
	selfId string
	// 署名用秘密鍵。
	key crypto.PrivateKey
	// 署名用秘密鍵の ID。
	kid string
	// 署名方式。
	alg string
	// ID トークンの有効期間。
	idTokExpiDur time.Duration

	base driver.TimeLimitedKeyValueStore
}

func (this *tokenContainerImpl) new(cod *code) (*token, error) {
	var tokId string
	for {
		if buff, err := util.SecureRandomString(this.idLen); err != nil {
			return nil, erro.Wrap(err)
		} else if val, _, err := this.base.Get(buff, nil); err != nil {
			return nil, erro.Wrap(err)
		} else if val == nil {
			// 昔発行した分とは重複しなかった。
			// 同時並列で発行している分と重複していない保証は無いが、まず大丈夫。
			tokId = buff
			break
		}
	}
	var refTok string
	log.Warn("Refresh token is not yet supported")
	jws := util.NewJws()
	jws.SetHeader(jwtAlg, this.alg)
	if this.kid != "" {
		jws.SetHeader(jwtKid, this.kid)
	}
	jws.SetClaim(clmIss, this.selfId)
	jws.SetClaim(clmSub, cod.accountId())
	jws.SetClaim(clmAud, cod.taId())
	now := time.Now()
	jws.SetClaim(clmExp, now.Add(this.idTokExpiDur).Unix())
	jws.SetClaim(clmIat, now.Unix())
	if !cod.authenticationDate().IsZero() {
		jws.SetClaim(clmAuthTim, cod.authenticationDate().Unix())
	}
	if cod.nonce() != "" {
		jws.SetClaim(clmNonc, cod.nonce())
	}
	if err := jws.Sign(map[string]crypto.PrivateKey{this.kid: this.key}); err != nil {
		return nil, erro.Wrap(err)
	}
	buff, err := jws.Encode()
	if err != nil {
		return nil, erro.Wrap(err)
	}
	idTok := string(buff)

	// アクセストークンが決まった。
	log.Debug("Token was generated")

	tok := newToken(
		tokId,
		cod.accountId(),
		cod.taId(),
		time.Now().Add(cod.expirationDuration()),
		refTok,
		cod.scopes(),
		idTok,
		cod.claims(),
	)
	if _, err := this.base.Put(tokId, tok, tok.expirationDate()); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Access token was published.")
	return tok, nil
}

func (this *tokenContainerImpl) get(tokId string) (*token, error) {
	val, _, err := this.base.Get(tokId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*token), nil
}
