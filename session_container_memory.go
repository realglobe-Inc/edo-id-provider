package main

import (
	"github.com/realglobe-Inc/edo-toolkit/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

type memorySessionContainer sessionContainerImpl

// スレッドセーフ。
func newMemorySessionContainer(minIdLen int, procId string, caStaleDur, caExpiDur time.Duration) sessionContainer {
	return &memorySessionContainer{
		driver.NewMemoryConcurrentVolatileKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
	}
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

func (this *memorySessionContainer) close() error {
	return ((*sessionContainerImpl)(this)).close()
}
