package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type memoryTokenContainer tokenContainerImpl

// スレッドセーフ。
func newMemoryTokenContainer(idLen int, expiDur, maxExpiDur, caStaleDur, caExpiDur time.Duration) *memoryTokenContainer {
	return (*memoryTokenContainer)(&tokenContainerImpl{
		idLen, expiDur, maxExpiDur,
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
	})
}

func (this *memoryTokenContainer) new(accId string, expiDur time.Duration) (*token, error) {
	return ((*tokenContainerImpl)(this)).new(accId, expiDur)
}

func (this *memoryTokenContainer) get(tokId string) (*token, error) {
	return ((*tokenContainerImpl)(this)).get(tokId)
}

func (this *memoryTokenContainer) add(tok *token) {
	((*tokenContainerImpl)(this)).base.(driver.TimeLimitedKeyValueStore).Put(tok.id(), tok, tok.expirationDate())
}
