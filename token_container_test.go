package main

import (
	"reflect"
	"testing"
	"time"
)

func testTokenContainer(t *testing.T, tokCont tokenContainer) {
	expiDur := 10 * time.Millisecond

	// 無い。
	if tok1, err := tokCont.get("ccccc"); err != nil {
		t.Fatal(err)
	} else if tok1 != nil {
		t.Error(tok1)
	}

	// 発行する。
	tok, err := tokCont.new("abcde", expiDur)
	if err != nil {
		t.Fatal(err)
	}

	// ある。
	tok2, err := tokCont.get(tok.Id)
	if err != nil {
		t.Fatal(err)
	} else if tok2 == nil || !reflect.DeepEqual(tok2, tok) {
		t.Error(tok2)
	}

	time.Sleep(tok.ExpiDate.Sub(time.Now()))

	// もう無い。
	if tok3, err := tokCont.get(tok.Id); err != nil {
		t.Fatal(err)
	} else if tok3 != nil {
		t.Error(tok3)
	}
}
