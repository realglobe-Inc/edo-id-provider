package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type sessionContainer interface {
	put(sess *session) error
	get(sessId string) (*session, error)
}

type sessionContainerImpl struct {
	base driver.TimeLimitedKeyValueStore
}

func (this *sessionContainerImpl) put(sess *session) error {
	if _, err := this.base.Put(sess.id(), sess, sess.expirationDate()); err != nil {
		return erro.Wrap(err)
	}

	return nil
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

type sessionContainerWrapper struct {
	// セッション番号の文字数。
	idLen int
	// デフォルトの有効期限。
	expiDur time.Duration

	sessionContainer
}

func (this *sessionContainerWrapper) put(sess *session) error {
	if sess.id() == "" {
		sessId, err := util.SecureRandomString(this.idLen)
		if err != nil {
			return erro.Wrap(err)
		}
		sess.setId(sessId)

		// セッション ID が決まった。
		log.Debug("Session ID " + mosaic(sess.id()) + " was generated")
	}

	sess.setExpirationDate(time.Now().Add(this.expiDur))
	return this.sessionContainer.put(sess)
}
