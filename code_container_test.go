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
	cod := newCode(
		codId,
		"account-id",
		"ta-id",
		"https://example.com/redirect/uri?a=b",
		time.Now().Add(time.Second),
		time.Second,
		nil,
		nil,
		"",
		time.Now())
	if err != nil {
		t.Fatal(err)
	}

	// 入れる。
	if err := codCont.put(cod); err != nil {
		t.Fatal(err)
	}

	// ある。
	if c, err := codCont.get(cod.id()); err != nil {
		t.Fatal(err)
	} else if c == nil {
		t.Error(c)
	} else if !reflect.DeepEqual(c, cod) {
		t.Error(c)
	}

	time.Sleep(cod.expirationDate().Sub(time.Now()) + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if c, err := codCont.get(cod.id()); err != nil {
		t.Fatal(err)
	} else if c != nil {
		t.Error(c)
	}
}
