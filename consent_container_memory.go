package main

import (
	"github.com/realglobe-Inc/edo-toolkit/driver"
	"net/url"
	"time"
)

// スレッドセーフ。
func newMemoryConsentContainer(staleDur, expiDur time.Duration) consentContainer {
	return &consentContainerImpl{
		driver.NewMemoryListedKeyValueStore(staleDur, expiDur),
		func(accId, taId string) string {
			return url.QueryEscape(accId) + "/" + url.QueryEscape(taId) // QueryEscape は '/' をエスケープするため。
		},
	}
}
