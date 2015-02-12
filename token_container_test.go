package main

import (
	"reflect"
	"testing"
	"time"
)

func testTokenContainer(t *testing.T, tokCont tokenContainer) {
	defer tokCont.close()

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
	for end := tok.expirationDate().Add(tokCont.(*tokenContainerImpl).savDur - time.Millisecond); ; // redis の粒度がミリ秒のため。
	{
		tk, err := tokCont.get(tok.id())
		if err != nil {
			t.Fatal(err)
		}
		cur := time.Now()

		// get と time.Now() の間に GC 等で時間が掛かることもあるため、
		// cur > end でも nil が返っているとは限らない。
		// cur <= end であれば非 nil が返らなければならない。

		if tk == nil {
			if cur.After(end) {
				break
			} else {
				t.Fatal(cur, end)
			}
		} else if tk.id() != tok.id() || tk.valid() {
			t.Error(tk, cur, end)
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
