package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type memoryCodeContainer codeContainerImpl

// スレッドセーフ。
func newMemoryCodeContainer(idLen int, expiDur, caStaleDur, caExpiDur time.Duration) *memoryCodeContainer {
	return (*memoryCodeContainer)(&codeContainerImpl{
		idLen, expiDur,
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
	})
}

func (this *memoryCodeContainer) new(accId, taId, rediUri string) (*code, error) {
	return ((*codeContainerImpl)(this)).new(accId, taId, rediUri)
}

func (this *memoryCodeContainer) get(codId string) (*code, error) {
	return ((*codeContainerImpl)(this)).get(codId)
}

func (this *memoryCodeContainer) add(cod *code) {
	((*codeContainerImpl)(this)).base.(driver.TimeLimitedKeyValueStore).Put(cod.Id, cod, time.Now().Add(((*codeContainerImpl)(this)).expiDur))
}
