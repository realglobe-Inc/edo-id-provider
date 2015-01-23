package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func newMemoryTokenContainer(minIdLen int, savDur, caStaleDur, caExpiDur time.Duration) tokenContainer {
	return &tokenContainerImpl{
		driver.NewMemoryTimeLimitedKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen),
		savDur,
	}
}
