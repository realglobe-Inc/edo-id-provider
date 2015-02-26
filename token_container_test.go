package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func testTokenContainer(t *testing.T, tokCont tokenContainer) {
	defer tokCont.close()

	savDur := tokCont.(*tokenContainerImpl).savDur

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
	bef := time.Now()
	if err := tokCont.put(tok); err != nil {
		t.Fatal(err)
	}
	diff := int64(time.Since(bef) / time.Nanosecond)

	// 無効になって消えるかどうか。
	for disap := tok.expirationDate().Add(savDur); ; {
		bef := time.Now()
		tk, err := tokCont.get(tok.id())
		if err != nil {
			t.Fatal(err)
		}
		aft := time.Now()

		// GC 等で時間が掛かることもあるため、aft > disap でも nil が返るとは限らない。
		// だが、aft <= disap であれば非 nil が返らなければならない。
		// 同様に、bef > disap であれば nil が返らなければならない。

		if aft.UnixNano() <= cutOff(disap.UnixNano(), 1e6)-diff { // redis の粒度がミリ秒のため。
			if tk == nil {
				t.Error(aft)
				t.Error(disap)
				return
			}
		} else if bef.UnixNano() > cutOff(disap.UnixNano(), 1e6)+1e6+diff { // redis の粒度がミリ秒のため。
			if tk != nil {
				t.Error(bef)
				t.Error(disap)
				return
			}
			// 消えた。
			return
		} else if tk == nil { // bef <= disap < aft
			// 消えた。
			return
		}

		bef = time.Now()
		ok := tk.valid()
		aft = time.Now()

		if !aft.After(exp) && !ok {
			t.Error(aft, exp)
			return
		} else if bef.After(exp) && ok {
			t.Error(bef, exp)
			return
		} else if !reflect.DeepEqual(tk, tok) {
			t.Error(fmt.Sprintf("%#v", tk))
			t.Error(fmt.Sprintf("%#v", tok))
			return
		}

		time.Sleep(time.Millisecond)
	}
}
