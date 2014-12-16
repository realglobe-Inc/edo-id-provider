package main

import (
	"reflect"
	"testing"
)

const (
	filePerm = 0644
)

const (
	testLabel = "edo-test"
)

var testAcc = &account{
	Id:     "abcde",
	Name:   "aaaaa",
	Passwd: "12345",
}

func testAccountContainer(t *testing.T, accCont accountContainer) {
	if acc, err := accCont.get(testAcc.Id); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(acc, testAcc) {
		t.Error(acc)
	}

	if acc, err := accCont.get(testAcc.Id + "a"); err != nil {
		t.Fatal(err)
	} else if acc != nil {
		t.Error(acc)
	}

	if acc, err := accCont.getByName(testAcc.Name); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(acc, testAcc) {
		t.Error(acc)
	}
	if acc, err := accCont.getByName(testAcc.Name + "a"); err != nil {
		t.Fatal(err)
	} else if acc != nil {
		t.Error(acc)
	}
}
