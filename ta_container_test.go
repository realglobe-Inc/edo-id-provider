package main

import (
	"reflect"
	"testing"
)

func testTaContainer(t *testing.T, taCont taContainer) {
	defer taCont.close()

	if ta_, err := taCont.get(testTa.id()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ta_, testTa) {
		t.Error(ta_, testTa)
	}

	if ta_, err := taCont.get(testTa.id() + "a"); err != nil {
		t.Fatal(err)
	} else if ta_ != nil {
		t.Error(ta_)
	}
}
