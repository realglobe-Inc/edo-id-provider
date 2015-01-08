package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
	"time"
)

func TestRedisTokenContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testTokenContainer(t, newRedisTokenContainer(10, "https://example.com", testPriKey, "", "RS256", time.Second, driver.NewRedisPool(redisAddr, 2, time.Second), testLabel, time.Second, time.Second))
}
