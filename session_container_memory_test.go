package main

import (
	logutil "github.com/realglobe-Inc/edo/util/log"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"testing"
)

func init() {
	logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
}

func TestMemorySessionContainer(t *testing.T) {
	// ////////////////////////////////
	// logutil.SetupConsole("github.com/realglobe-Inc", level.ALL)
	// defer logutil.SetupConsole("github.com/realglobe-Inc", level.OFF)
	// ////////////////////////////////
	testSessionContainer(t, newMemorySessionContainer(10, "", testStaleDur, testCaExpiDur))
}
