package main

import (
	"testing"
	"time"
)

func TestMemoryTokenContainer(t *testing.T) {
	testTokenContainer(t, newMemoryTokenContainer(10, "https://example.com", testPriKey, "", "RS256", time.Second, time.Second, time.Second))
}
