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
	tok, err := tokCont.new(newCode("abcde", "account", "ta", "redirect_uri", time.Now().Add(time.Second), expiDur, nil, nil, "nonce", time.Time{}))
	if err != nil {
		t.Fatal(err)
	}

	// ある。
	tok2, err := tokCont.get(tok.id())
	if err != nil {
		t.Fatal(err)
	} else if tok2 == nil {
		t.Error(tok2)
	} else if !reflect.DeepEqual(tok2, tok) {
		t.Error(tok2)
	}

	time.Sleep(tok.expirationDate().Sub(time.Now()) + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if tok3, err := tokCont.get(tok.id()); err != nil {
		t.Fatal(err)
	} else if tok3 != nil {
		t.Error(tok3)
	}
}
