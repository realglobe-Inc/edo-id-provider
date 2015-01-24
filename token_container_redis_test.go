package main

import (
	"testing"
)

func TestRedisTokenContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testTokenContainer(t, newRedisTokenContainer(10, "https://example.com", testPriKey, "", "RS256", testSavDur, testRedisPool, testLabel, testStaleDur, testCaExpiDur))
}
