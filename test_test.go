package main

import (
	"time"
)

// テスト用の諸々。

const (
	filePerm = 0644
)

const (
	testLabel = "edo-test"

	testSavDur    = 15 * time.Millisecond
	testStaleDur  = 5 * time.Millisecond
	testCaExpiDur = time.Millisecond

	testTicDur = 10 * time.Millisecond
)
