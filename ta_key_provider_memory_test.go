package main

import (
	"testing"
)

func TestMemoryTaKeyProvider(t *testing.T) {
	reg := NewMemoryTaKeyProvider(0)
	reg.AddServiceKey(testServUuid, testPublicKey)
	testTaKeyProvider(t, reg)
}

func TestMemoryTaKeyProviderStamp(t *testing.T) {
	reg := NewMemoryTaKeyProvider(0)
	reg.AddServiceKey(testServUuid, testPublicKey)
	testTaKeyProvider(t, reg)
}
