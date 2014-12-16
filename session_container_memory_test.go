package main

import (
	"testing"
	"time"
)

func TestMemorySessionContainer(t *testing.T) {
	testSessionContainer(t, newMemorySessionContainer(10, 10*time.Millisecond, time.Second, time.Second))
}
