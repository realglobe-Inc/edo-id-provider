package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
	"time"
)

func TestRedisCodeContainer(t *testing.T) {
	if redisAddr == "" {
		t.SkipNow()
	}
	testCodeContainer(t, newRedisCodeContainer(10, 10*time.Millisecond, driver.NewRedisPool(redisAddr, 2, time.Second), testLabel, time.Second, time.Second))
}
