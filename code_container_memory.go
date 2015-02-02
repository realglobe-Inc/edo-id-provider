package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func newMemoryCodeContainer(minIdLen int, procId string, savDur, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &codeContainerImpl{
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
	}
}
