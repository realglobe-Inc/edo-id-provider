package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type memorySessionContainer sessionContainerImpl

// スレッドセーフ。
func newMemorySessionContainer(idLen int, expiDur, caStaleDur, caExpiDur time.Duration) *memorySessionContainer {
	return (*memorySessionContainer)(&sessionContainerImpl{
		idLen, expiDur,
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
	})
}

func (this *memorySessionContainer) new(accId string, expiDur time.Duration) (*session, error) {
	return ((*sessionContainerImpl)(this)).new(accId, expiDur)
}

func (this *memorySessionContainer) get(sessId string) (*session, error) {
	return ((*sessionContainerImpl)(this)).get(sessId)
}

func (this *memorySessionContainer) update(sess *session) error {
	return ((*sessionContainerImpl)(this)).update(sess)
}

func (this *memorySessionContainer) add(sess *session) {
	((*sessionContainerImpl)(this)).base.(driver.TimeLimitedKeyValueStore).Put(sess.Id, sess, sess.ExpiDate)
}
