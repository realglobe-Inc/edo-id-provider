package main

import (
	"testing"
)

func TestMemoryTokenContainer(t *testing.T) {
	testTokenContainer(t, newMemoryTokenContainer(10, "https://example.com", testPriKey, "", "RS256", testSavDur, testStaleDur, testCaExpiDur))
}
