package main

import (
	"github.com/realglobe-Inc/edo-lib/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

type tokenContainer interface {
	newId() (id string, err error)
	get(tokId string) (*token, error)
	put(tok *token) error

	close() error
}

type tokenContainerImpl struct {
	base driver.VolatileKeyValueStore

	idGenerator
	// 有効期限が切れてからも保持する期間。
	savDur time.Duration
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

func (this *tokenContainerImpl) put(tok *token) error {
	if _, err := this.base.Put(tok.id(), tok, tok.expirationDate().Add(this.savDur)); err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (this *tokenContainerImpl) close() error {
	return this.base.Close()
}
