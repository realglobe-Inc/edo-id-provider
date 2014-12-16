package main

import (
	"testing"
)

func TestMemoryTaContainer(t *testing.T) {
	taCont := newMemoryTaContainer(0, 0)
	taCont.add(testTa)
	testTaContainer(t, taCont)
}
