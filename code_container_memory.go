package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"time"
)

// スレッドセーフ。
func newMemoryCodeContainer(minIdLen int, procId string, savDur, ticExpDur, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &codeContainerImpl{
		driver.NewMemoryConcurrentVolatileKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
		ticExpDur,
	}
}
