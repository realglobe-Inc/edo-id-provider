package main

import (
	"testing"
)

func TestMemoryCodeContainer(t *testing.T) {
	testCodeContainer(t, newMemoryCodeContainer(10, "", testSavDur, testTicDur, testStaleDur, testCaExpiDur))
}
