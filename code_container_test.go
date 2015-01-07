package main

import (
	"reflect"
	"testing"
	"time"
)

func testCodeContainer(t *testing.T, codCont codeContainer) {
	// 無い。
	if cod1, err := codCont.get("ccccc"); err != nil {
		t.Fatal(err)
	} else if cod1 != nil {
		t.Error(cod1)
	}

	// 発行する。
	cod, err := codCont.new("abcde", "ABCDE", "https://example.com/a/b/c?a=b", time.Second, nil, "", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	// ある。
	if cod2, err := codCont.get(cod.id()); err != nil {
		t.Fatal(err)
	} else if cod2 == nil {
		t.Error(cod2)
	} else if !reflect.DeepEqual(cod2, cod) {
		t.Error(cod2)
	}

	time.Sleep(cod.expirationDate().Sub(time.Now()))

	// もう無い。
	if cod3, err := codCont.get(cod.id()); err != nil {
		t.Fatal(err)
	} else if cod3 != nil {
		t.Error(cod3)
	}
}
