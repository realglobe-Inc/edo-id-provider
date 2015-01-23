package main

import (
	"encoding/base64"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"math/big"
	"math/rand"
	"sync/atomic"
	"time"
)

type sessionContainer interface {
	newId() (id string, err error)
	put(sess *session) error
	get(sessId string) (*session, error)
}

type sessionContainerImpl struct {
	base driver.TimeLimitedKeyValueStore

	// セッション文字数の下界。
	minIdLen int
	// インスタンス内でのセッション被りを防ぐための通し番号。
	// 別インスタンスは保証できない。
	ser int64
}

func newSessionContainerImpl(base driver.TimeLimitedKeyValueStore, minIdLen int) *sessionContainerImpl {
	return &sessionContainerImpl{
		base:     base,
		minIdLen: minIdLen,
		ser:      rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
	}
}

func (this *sessionContainerImpl) newId() (id string, err error) {
	id, err = util.SecureRandomString(this.minIdLen)
	if err != nil {
		return "", erro.Wrap(err)
	}
	buff := big.NewInt(atomic.AddInt64(&this.ser, 1)).Bytes()
	id += base64.URLEncoding.EncodeToString(buff)[:(len(buff)*8+5)/6]
	return id, nil
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
