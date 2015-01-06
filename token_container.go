package main

import (
	"encoding/base64"
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

	base driver.TimeLimitedKeyValueStore
}

func (this *tokenContainerImpl) new(cod *code) (*token, error) {
	var tokId string
	for {
		buff, err := util.SecureRandomBytes(this.idLen * 6 / 8)
		if err != nil {
			return nil, erro.Wrap(err)
		}

		tokId = base64.URLEncoding.EncodeToString(buff)
		if val, _, err := this.base.Get(tokId, nil); err != nil {
			return nil, erro.Wrap(err)
		} else if val == nil {
			// 昔発行した分とは重複しなかった。
			// 同時並列で発行している分と重複していない保証は無いが、まず大丈夫。
			break
		}
	}

	// アクセストークンが決まった。
	log.Debug("Token was generated")

	tok := newToken(tokId, cod.accountId(), time.Now().Add(cod.expirationDuration()))
	if _, err := this.base.Put(tokId, tok, tok.ExpiDate); err != nil {
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
