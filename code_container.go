package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type codeContainer interface {
	// expiDur は後で発行するアクセストークンの有効期間。
	new(accId,
		taId,
		rediUri string,
		expiDur time.Duration,
		scops,
		clms map[string]bool,
		nonc string,
		authDate time.Time) (*code, error)
	get(codId string) (*code, error)
}

type codeContainerImpl struct {
	// 認可コードの識別子部分の文字数。
	idLen int
	// 認可コードの有効期間。
	expiDur time.Duration
	// 自分の IdP としての ID。
	selfId string

	base driver.TimeLimitedKeyValueStore
}

func (this *codeContainerImpl) new(accId,
	taId,
	rediUri string,
	expiDur time.Duration,
	scops,
	clms map[string]bool,
	nonc string,
	authDate time.Time) (*code, error) {

	var codId string
	for {
		jti, err := util.SecureRandomString(this.idLen)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		jws := util.NewJws()
		jws.SetHeader(jwtAlg, algNone)
		jws.SetClaim(clmJti, jti)
		jws.SetClaim(clmIss, this.selfId)
		if err := jws.Sign(nil); err != nil {
			return nil, erro.Wrap(err)
		}
		buff, err := jws.Encode()
		if err != nil {
			return nil, erro.Wrap(err)
		}
		codId = string(buff)
		if val, _, err := this.base.Get(codId, nil); err != nil {
			return nil, erro.Wrap(err)
		} else if val == nil {
			// 昔発行した分とは重複しなかった。
			// 同時並列で発行している分と重複していない保証は無いが、まず大丈夫。
			break
		}

	}

	// コードが決まった。
	log.Debug("Code was generated.")

	cod := newCode(codId, accId, taId, rediUri, time.Now().Add(this.expiDur), expiDur, scops, clms, nonc, authDate)
	if _, err := this.base.Put(codId, cod, cod.expirationDate()); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Code was published.")
	return cod, nil
}

func (this *codeContainerImpl) get(codId string) (*code, error) {
	val, _, err := this.base.Get(codId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*code), nil
}
