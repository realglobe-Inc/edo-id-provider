package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
)

type sessionContainer interface {
	newId() (id string, err error)
	put(sess *session) error
	get(sessId string) (*session, error)
}

type sessionContainerImpl struct {
	base driver.TimeLimitedKeyValueStore

	idGenerator
}

func newSessionContainerImpl(base driver.TimeLimitedKeyValueStore, minIdLen int) *sessionContainerImpl {
	return &sessionContainerImpl{
		base:        base,
		idGenerator: newIdGenerator(minIdLen),
	}
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
