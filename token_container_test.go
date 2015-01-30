package main

import (
	"reflect"
	"testing"
	"time"
)

func testTokenContainer(t *testing.T, tokCont tokenContainer) {
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
	exp := time.Now().Add(testTokExpiDur)
	tok := newToken(id,
		"testaccount",
		"testta",
		"testcode",
		"",
		exp,
		nil,
		nil,
		"")

	// 入れる。
	if err := tokCont.put(tok); err != nil {
		t.Fatal(err)
	}

	// 有効。
	for cur := time.Now(); cur.Before(exp); cur = time.Now() {
		if tk, err := tokCont.get(tok.id()); err != nil {
			t.Fatal(err)
		} else if tk == nil {
			t.Fatal(cur, exp)
		} else if !reflect.DeepEqual(tk, tok) {
			t.Error(tk, cur, exp)
		}

		time.Sleep(testTokExpiDur / 4)
	}

	// 無効。
	for cur, end := time.Now(), tok.expirationDate().Add(tokCont.(*tokenContainerImpl).savDur-time.Millisecond); // redis の粒度がミリ秒のため。
	cur.Before(end); cur = time.Now() {
		if tk, err := tokCont.get(tok.id()); err != nil {
			t.Fatal(err)
		} else if tk == nil {
			t.Fatal(cur, exp)
		} else if tk.id() != tok.id() || tk.valid() {
			t.Error(tk, cur, exp)
		}

		time.Sleep(tokCont.(*tokenContainerImpl).savDur / 4)
	}

	time.Sleep(time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if tk, err := tokCont.get(tok.id()); err != nil {
		t.Fatal(err)
	} else if tk != nil {
		t.Error(tk)
	}
}
