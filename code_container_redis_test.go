package main

import (
	"testing"
)

func TestRedisCodeContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testCodeContainer(t, newRedisCodeContainer(10, "", testSavDur, testTicDur, testRedisPool, testLabel, testStaleDur, testCaExpiDur))
}
