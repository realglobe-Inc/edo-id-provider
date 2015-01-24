package main

import (
	"testing"
)

func TestMemoryTaContainer(t *testing.T) {
	taCont := newMemoryTaContainer(testStaleDur, testCaExpiDur)
	taCont.add(testTa)
	testTaContainer(t, taCont)
}
