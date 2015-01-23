package main

import (
	"testing"
	"time"
)

func testSessionContainer(t *testing.T, sessCont sessionContainer) {
	expiDur := 20 * time.Millisecond

	// 無い。
	if se, err := sessCont.get("ccccc"); err != nil {
		t.Fatal(err)
	} else if se != nil {
		t.Error(se)
	}

	// 発行する。
	sess := newSession()
	id, err := sessCont.newId()
	if err != nil {
		t.Fatal(err)
	}

	sess.setId(id)
	sess.setExpirationDate(time.Now().Add(expiDur))
	if err := sessCont.put(sess); err != nil {
		t.Fatal(err)
	}

	// ある。
	for i := 0; i < 4; i++ {
		se, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		} else if se == nil {
			t.Fatal(i, se)
		} else if se.id() != sess.id() {
			t.Error(i, se)
		}
		s := *se
		s.setExpirationDate(time.Now().Add(expiDur))
		if err := sessCont.put(&s); err != nil {
			t.Fatal(err)
		}
		sess = &s
		time.Sleep(expiDur / 2)
	}

	time.Sleep(expiDur/2 + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if se, err := sessCont.get(sess.id()); err != nil {
		t.Fatal(err)
	} else if se != nil {
		t.Error(se)
	}
}
