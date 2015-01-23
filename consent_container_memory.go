package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func newMemoryConsentContainer(staleDur, expiDur time.Duration) consentContainer {
	return &consentContainerImpl{
		driver.NewMemoryListedKeyValueStore(staleDur, expiDur),
		func(accId, taId string) string {
			return accId + "/" + taId
		},
	}
}
