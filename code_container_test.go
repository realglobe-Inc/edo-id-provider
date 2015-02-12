package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func testCodeContainer(t *testing.T, codCont codeContainer) {
	defer codCont.close()

	var savDur time.Duration
	switch c := codCont.(type) {
	case *codeContainerImpl:
		savDur = c.savDur
	case *memoryCodeContainer:
		savDur = c.savDur
	default:
		t.Fatal("unknown code container")
	}

	// 無い。
	if c, err := codCont.get("ccccc"); err != nil {
		t.Fatal(err)
	} else if c != nil {
		t.Error(c)
	}

	// 発行する。
	codId, err := codCont.newId()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	exp := now.Add(testCodExpiDur)
	cod := newCode(
		codId,
		"account-id",
		"ta-id",
		"https://example.com/redirect/uri?a=b",
		exp,
		testTokExpiDur,
		nil,
		nil,
		"",
		now)
	if err != nil {
		t.Fatal(err)
	}

	// 入れる。
	if err := codCont.put(cod); err != nil {
		t.Fatal(err)
	}

	// 有効。
	for cur := time.Now(); cur.Before(exp); cur = time.Now() {
		if c, err := codCont.get(cod.id()); err != nil {
			t.Fatal(err)
		} else if c == nil {
			t.Fatal(cur, exp)
		} else if !reflect.DeepEqual(c, cod) {
			t.Error(fmt.Sprintf("%#v", c))
			t.Error(fmt.Sprintf("%#v", cod))
		}

		time.Sleep(testCodExpiDur / 4)
	}

	// 無効。
	for end := cod.expirationDate().Add(savDur - time.Millisecond); ; // redis の粒度がミリ秒のため。
	{
		c, err := codCont.get(cod.id())
		if err != nil {
			t.Fatal(err)
		}
		cur := time.Now()

		// get と time.Now() の間に GC 等で時間が掛かることもあるため、
		// cur > end でも nil が返っているとは限らない。
		// cur <= end であれば非 nil が返らなければならない。

		if c == nil {
			if cur.After(end) {
				break
			} else {
				t.Fatal(cur, end)
			}
		} else if c.id() != cod.id() || c.valid() {
			t.Error(c, cur, end)
		}

		time.Sleep(savDur / 4)
	}

	time.Sleep(time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if c, err := codCont.get(cod.id()); err != nil {
		t.Fatal(err)
	} else if c != nil {
		t.Error(c)
	}
}
