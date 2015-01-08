package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
	"time"
)

func TestRedisSessionContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testSessionContainer(t, newRedisSessionContainer(10, 10*time.Millisecond, driver.NewRedisPool(redisAddr, 2, time.Second), testLabel, time.Second, time.Second))
}
