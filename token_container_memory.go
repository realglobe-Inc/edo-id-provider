package main

import (
	"github.com/realglobe-Inc/edo-lib/driver"
	"time"
)

// スレッドセーフ。
func newMemoryTokenContainer(minIdLen int, procId string, savDur, caStaleDur, caExpiDur time.Duration) tokenContainer {
	return &tokenContainerImpl{
		driver.NewMemoryConcurrentVolatileKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
	}
}
