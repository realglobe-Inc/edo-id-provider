package main

import (
	"testing"
	"time"
)

func TestMemoryCodeContainer(t *testing.T) {
	testCodeContainer(t, newMemoryCodeContainer(10, 10*time.Millisecond, "https://example.com", time.Second, time.Second))
}
