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
	id, err := tokCont.newId()
	if err != nil {
		t.Fatal(err)
	}
	tok := newToken(id,
		"testaccount",
		"testta",
		"testcode",
		"",
		time.Now().Add(expiDur),
		nil,
		nil,
		"")

	// 入れる。
	if err := tokCont.put(tok); err != nil {
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

	time.Sleep(tok.expirationDate().Add(tokCont.(*tokenContainerImpl).savDur).Sub(time.Now()) + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if tk, err := tokCont.get(tok.id()); err != nil {
		t.Fatal(err)
	} else if tk != nil {
		t.Error(tk)
	}
}
