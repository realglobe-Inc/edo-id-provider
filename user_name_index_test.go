package main

import (
	"testing"
)

const (
	testUsrName = "test-user-no-namae"
)

func testUserNameIndex(t *testing.T, reg UserNameIndex) {
	usrUuid1, _, err := reg.UserUuid(testUsrName, nil)
	if err != nil {
		t.Fatal(err)
	} else if usrUuid1 != testUsrUuid {
		t.Error(usrUuid1)
	}

	usrUuid2, _, err := reg.UserUuid(testUsrName+"_d", nil)
	if err != nil {
		t.Fatal(err)
	} else if usrUuid2 != "" {
		t.Error(usrUuid2)
	}
}
