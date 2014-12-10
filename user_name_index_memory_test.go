package main

import (
	"testing"
)

func TestMemoryUserNameIndex(t *testing.T) {
	reg := NewMemoryUserNameIndex(0)
	reg.AddUserUuid(testUsrName, testUsrUuid)
	testUserNameIndex(t, reg)
}
