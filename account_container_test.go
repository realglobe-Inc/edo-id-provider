package main

import (
	"testing"
)

const (
	filePerm = 0644
)

const (
	testLabel = "edo-test"
)

var testAcc = newAccount(map[string]interface{}{
	"id":     "abcde",
	"name":   "aaaaa",
	"passwd": "12345",
})

func testAccountContainer(t *testing.T, accCont accountContainer) {
	if acc, err := accCont.get(testAcc.id()); err != nil {
		t.Fatal(err)
	} else if acc.id() != testAcc.id() {
		t.Error(acc)
	} else if acc.name() != testAcc.name() {
		t.Error(acc)
	} else if acc.password() != testAcc.password() {
		t.Error(acc)
	}

	if acc, err := accCont.get(testAcc.id() + "a"); err != nil {
		t.Fatal(err)
	} else if acc != nil {
		t.Error(acc)
	}

	if acc, err := accCont.getByName(testAcc.name()); err != nil {
		t.Fatal(err)
	} else if acc.id() != testAcc.id() {
		t.Error(acc)
	} else if acc.name() != testAcc.name() {
		t.Error(acc)
	} else if acc.password() != testAcc.password() {
		t.Error(acc)
	}

	if acc, err := accCont.getByName(testAcc.name() + "a"); err != nil {
		t.Fatal(err)
	} else if acc != nil {
		t.Error(acc)
	}
}
