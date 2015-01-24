package main

import (
	"testing"
)

func TestMemoryConsentContainer(t *testing.T) {
	testConsentContainer(t, newMemoryConsentContainer(testStaleDur, testCaExpiDur))
}
