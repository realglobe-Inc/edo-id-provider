package main

import (
	"encoding/base64"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type session struct {
	// セッション番号。
	Id string `json:"id"`
	// 対象のアカウント。
	AccId string `json:"account_id"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`
}

type sessionContainer interface {
	new(accId string, expiDur time.Duration) (*session, error)
	get(sessId string) (*session, error)
	update(sess *session) error
}

type sessionContainerImpl struct {
	// セッション番号の文字数。
	idLen int
	// デフォルトの有効期限。
	expiDur time.Duration

	base driver.TimeLimitedKeyValueStore
}

func (this *sessionContainerImpl) new(accId string, expiDur time.Duration) (*session, error) {
	var sessId string
	for {
		buff, err := util.SecureRandomBytes(this.idLen * 6 / 8)
		if err != nil {
			return nil, erro.Wrap(err)
		}

		sessId = base64.URLEncoding.EncodeToString(buff)
		if val, _, err := this.base.Get(sessId, nil); err != nil {
			return nil, erro.Wrap(err)
		} else if val == nil {
			// 昔発行した分とは重複しなかった。
			// 同時並列で発行している分と重複していない保証は無いが、まず大丈夫。
			break
		}
	}

	// セッション ID が決まった。
	log.Debug("Session ID was generated")

	if expiDur == 0 || expiDur > this.expiDur {
		expiDur = this.expiDur
	}

	sess := &session{sessId, accId, time.Now().Add(expiDur)}
	if _, err := this.base.Put(sessId, sess, sess.ExpiDate); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Session was published.")
	return sess, nil
}

func (this *sessionContainerImpl) get(sessId string) (*session, error) {
	val, _, err := this.base.Get(sessId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*session), nil
}

func (this *sessionContainerImpl) update(sess *session) error {
	if _, err := this.base.Put(sess.Id, sess, sess.ExpiDate); err != nil {
		return erro.Wrap(err)
	}
	return nil
}
