package main

import (
	"testing"
)

func TestMemoryAccountContainer(t *testing.T) {
	accCont := newMemoryAccountContainer(0, 0)
	accCont.add(testAcc)
	testAccountContainer(t, accCont)
}
