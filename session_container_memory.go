package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type memorySessionContainer sessionContainerImpl

// スレッドセーフ。
func newMemorySessionContainer(minIdLen int, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return newSessionContainerImpl(
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
		minIdLen)
}

func (this *memorySessionContainer) put(sess *session) error {
	return ((*sessionContainerImpl)(this)).put(sess.copy())
}

func (this *memorySessionContainer) get(sessId string) (*session, error) {
	sess, err := ((*sessionContainerImpl)(this)).get(sessId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if sess == nil {
		return nil, nil
	}
	return sess.copy(), nil
}
