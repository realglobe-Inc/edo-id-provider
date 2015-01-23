package main

import (
	"testing"
)

var testAcc = newAccount(map[string]interface{}{
	"id":       "testaccount",
	"username": "aaaaa",
	"password": "12345",
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
