package main

import (
	"encoding/base64"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type code struct {
	// 認可コード。
	Id string `json:"id"`
	// 発行したアカウント。
	AccId string `json:"account_id"`
	// 発行先 TA。
	TaId string `json:"ta_id"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`
}

type codeContainer interface {
	new(accId, taId string) (*code, error)
	get(codId string) (*code, error)
}

type codeContainerImpl struct {
	// 認可コードの文字数。
	idLen int
	// 有効期限。
	expiDur time.Duration

	base driver.TimeLimitedKeyValueStore
}

func (this *codeContainerImpl) new(accId, taId string) (*code, error) {
	var codId string
	for {
		buff, err := util.SecureRandomBytes(this.idLen * 6 / 8)
		if err != nil {
			return nil, erro.Wrap(err)
		}

		codId = base64.URLEncoding.EncodeToString(buff)
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

	cod := &code{codId, accId, taId, time.Now().Add(this.expiDur)}
	if _, err := this.base.Put(codId, cod, cod.ExpiDate); err != nil {
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
