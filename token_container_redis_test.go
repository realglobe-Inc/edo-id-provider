package main

import (
	"testing"
)

func TestRedisTokenContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testTokenContainer(t, newRedisTokenContainer(10, testSavDur, testRedisPool, testLabel, testStaleDur, testCaExpiDur))
}
