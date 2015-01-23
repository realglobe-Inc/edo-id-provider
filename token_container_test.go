package main

import (
	"reflect"
	"testing"
	"time"
)

func testTokenContainer(t *testing.T, tokCont tokenContainer) {
	expiDur := 10 * time.Millisecond

	// 無い。
	if tk, err := tokCont.get("ccccc"); err != nil {
		t.Fatal(err)
	} else if tk != nil {
		t.Error(tk)
	}

	// 発行する。
	tok, err := tokCont.new(newCode("abcde", "account", "ta", "redirect_uri", time.Now().Add(time.Second), expiDur, nil, nil, "nonce", time.Time{}))
	if err != nil {
		t.Fatal(err)
	}

	// ある。
	if tk, err := tokCont.get(tok.id()); err != nil {
		t.Fatal(err)
	} else if tk == nil {
		t.Error(tk)
	} else if !reflect.DeepEqual(tk, tok) {
		t.Error(tk)
	}

	time.Sleep(tok.expirationDate().Sub(time.Now()) + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if tk, err := tokCont.get(tok.id()); err != nil {
		t.Fatal(err)
	} else if tk != nil {
		t.Error(tk)
	}
}
