package main

import (
	"testing"
)

func TestRedisCodeContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testCodeContainer(t, newRedisCodeContainer(10, testSavDur, testRedisPool, testLabel, testStaleDur, testCaExpiDur))
}
