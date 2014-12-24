package main

import (
	"testing"
	"time"
)

func testSessionContainer(t *testing.T, sessCont sessionContainer) {
	expiDur := 10 * time.Millisecond

	// 無い。
	if sess1, err := sessCont.get("ccccc"); err != nil {
		t.Fatal(err)
	} else if sess1 != nil {
		t.Error(sess1)
	}

	// 発行する。
	sess := newSession()
	if err := sessCont.put(sess); err != nil {
		t.Fatal(err)
	}

	// ある。
	for i := 0; i < 4; i++ {
		sess2, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		} else if sess2 == nil || sess2.id() != sess.id() {
			t.Error(i, sess2)
		}
		s := *sess2
		s.setExpirationDate(time.Now().Add(expiDur))
		if err := sessCont.put(&s); err != nil {
			t.Fatal(err)
		}
		sess = &s
		time.Sleep(expiDur / 2)
	}

	time.Sleep(expiDur / 2)

	// もう無い。
	if sess3, err := sessCont.get(sess.id()); err != nil {
		t.Fatal(err)
	} else if sess3 != nil {
		t.Error(sess3)
	}
}