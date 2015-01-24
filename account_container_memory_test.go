package main

import (
	"testing"
)

func TestMemoryAccountContainer(t *testing.T) {
	accCont := newMemoryAccountContainer(testStaleDur, testCaExpiDur)
	accCont.add(testAcc)
	testAccountContainer(t, accCont)
}
