package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func newMemoryTokenContainer(minIdLen int, procId string, savDur, caStaleDur, caExpiDur time.Duration) tokenContainer {
	return &tokenContainerImpl{
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
	}
}
