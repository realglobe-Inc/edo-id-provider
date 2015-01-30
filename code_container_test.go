package main

import (
	"reflect"
	"testing"
	"time"
)

func testCodeContainer(t *testing.T, codCont codeContainer) {
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
			t.Error(c, cur, exp)
		}

		time.Sleep(testCodExpiDur / 4)
	}

	// 無効。
	for cur, end := time.Now(), cod.expirationDate().Add(codCont.(*codeContainerImpl).savDur-time.Millisecond); // redis の粒度がミリ秒のため。
	cur.Before(end); cur = time.Now() {
		if c, err := codCont.get(cod.id()); err != nil {
			t.Fatal(err)
		} else if c == nil {
			t.Fatal(cur, exp)
		} else if c.id() != cod.id() || c.valid() {
			t.Error(c, cur, exp)
		}

		time.Sleep(codCont.(*codeContainerImpl).savDur / 4)
	}

	time.Sleep(time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if c, err := codCont.get(cod.id()); err != nil {
		t.Fatal(err)
	} else if c != nil {
		t.Error(c)
	}
}
