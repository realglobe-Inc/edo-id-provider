package main

import (
	"testing"
)

func TestMemoryTokenContainer(t *testing.T) {
	testTokenContainer(t, newMemoryTokenContainer(10, "", testSavDur, testStaleDur, testCaExpiDur))
}
