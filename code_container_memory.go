package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type memoryCodeContainer codeContainerImpl

// スレッドセーフ。
func newMemoryCodeContainer(idLen int, expiDur time.Duration, selfId string, caStaleDur, caExpiDur time.Duration) *memoryCodeContainer {
	return (*memoryCodeContainer)(&codeContainerImpl{
		idLen, expiDur, selfId,
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
	})
}

func (this *memoryCodeContainer) new(accId, taId, rediUri string, expiDur time.Duration, scops map[string]bool, nonc string, authDate time.Time) (*code, error) {
	return ((*codeContainerImpl)(this)).new(accId, taId, rediUri, expiDur, scops, nonc, authDate)
}

func (this *memoryCodeContainer) get(codId string) (*code, error) {
	return ((*codeContainerImpl)(this)).get(codId)
}

func (this *memoryCodeContainer) add(cod *code) {
	((*codeContainerImpl)(this)).base.(driver.TimeLimitedKeyValueStore).Put(cod.id(), cod, time.Now().Add(((*codeContainerImpl)(this)).expiDur))
}
