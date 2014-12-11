package main

import (
	"testing"
)

func TestMemoryTaExplorer(t *testing.T) {
	reg := NewMemoryTaExplorer(0, 0)
	reg.SetServiceUuids(map[string]string{testUri: testServUuid})
	testTaExplorer(t, reg)
}

func TestMemoryTaExplorerStamp(t *testing.T) {
	reg := NewMemoryTaExplorer(0, 0)
	reg.SetServiceUuids(map[string]string{testUri: testServUuid})
	testTaExplorerStamp(t, reg)
}
