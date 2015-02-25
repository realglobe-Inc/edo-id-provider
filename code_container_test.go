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
	bef := time.Now()
	if err := codCont.put(cod); err != nil {
		t.Fatal(err)
	}
	diff := int64(time.Since(bef) / time.Nanosecond)

	// 無効になって消えるかどうか。
	disap := cod.expirationDate().Add(savDur)
	for deadline := disap.Add(time.Second); ; {
		bef := time.Now()
		c, err := codCont.get(cod.id())
		if err != nil {
			t.Fatal(err)
		}
		aft := time.Now()

		// GC 等で時間が掛かることもあるため、aft > disap でも nil が返るとは限らない。
		// だが、aft <= disap であれば非 nil が返らなければならない。
		// 同様に、bef > disap であれば nil が返らなければならない。

		if aft.UnixNano() <= cutOff(disap.UnixNano(), 1e6)-diff { // redis の粒度がミリ秒のため。
			if c == nil {
				t.Error(aft)
				t.Error(disap)
				return
			}
		} else if bef.UnixNano() > cutOff(disap.UnixNano(), 1e6)+1e6+diff { // redis の粒度がミリ秒のため。
			if c != nil {
				t.Error(bef)
				t.Error(disap)
				return
			}
			// 消えた。
			return
		} else if c == nil { // bef <= disap < aft
			// 消えた。
			return
		}

		bef = time.Now()
		ok := c.valid()
		aft = time.Now()

		if !aft.After(exp) && !ok {
			t.Error(aft)
			t.Error(exp)
			return
		} else if bef.After(exp) && ok {
			t.Error(bef)
			t.Error(exp)
			return
		} else if !reflect.DeepEqual(c, cod) {
			t.Error(fmt.Sprintf("%#v", c))
			t.Error(fmt.Sprintf("%#v", cod))
			return
		}

		if aft.After(deadline) {
			t.Error("too late")
			return
		}
		time.Sleep(time.Millisecond)
	}
}
