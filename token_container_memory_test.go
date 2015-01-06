package main

import (
	"testing"
	"time"
)

func TestMemoryTokenContainer(t *testing.T) {
	testTokenContainer(t, newMemoryTokenContainer(10, time.Second, time.Second))
}
