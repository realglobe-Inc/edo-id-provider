package main

import (
	"testing"
)

func TestMemoryUserAttributeRegistry(t *testing.T) {
	reg := NewMemoryUserAttributeRegistry(0)
	reg.AddUserAttribute(testUsrUuid, testAttrName, testAttr)
	testUserAttributeRegistry(t, reg)
}
