package main

import (
	"testing"
)

func TestRedisSessionContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testSessionContainer(t, newRedisSessionContainer(10, "", testRedisPool, testLabel, testStaleDur, testCaExpiDur))
}
