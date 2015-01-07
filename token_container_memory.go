package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

type memoryTokenContainer tokenContainerImpl

// スレッドセーフ。
func newMemoryTokenContainer(idLen int, selfId string, key crypto.PrivateKey, kid, alg string, idTokExpiDur time.Duration,
	caStaleDur, caExpiDur time.Duration) *memoryTokenContainer {
	return (*memoryTokenContainer)(&tokenContainerImpl{
		idLen, selfId, key, kid, alg, idTokExpiDur,
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
	})
}

func (this *memoryTokenContainer) new(cod *code) (*token, error) {
	return ((*tokenContainerImpl)(this)).new(cod)
}

func (this *memoryTokenContainer) get(tokId string) (*token, error) {
	return ((*tokenContainerImpl)(this)).get(tokId)
}

func (this *memoryTokenContainer) add(tok *token) {
	((*tokenContainerImpl)(this)).base.(driver.TimeLimitedKeyValueStore).Put(tok.id(), tok, tok.expirationDate())
}
