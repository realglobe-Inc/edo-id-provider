package main

import (
	"reflect"
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
	sess, err := sessCont.new("abcde", expiDur)
	if err != nil {
		t.Fatal(err)
	}

	// ある。
	for i := 0; i < 4; i++ {
		sess2, err := sessCont.get(sess.Id)
		if err != nil {
			t.Fatal(err)
		} else if sess2 == nil || !reflect.DeepEqual(sess2, sess) {
			t.Error(i, sess2)
		}
		s := *sess2
		s.ExpiDate = time.Now().Add(expiDur)
		if err := sessCont.update(&s); err != nil {
			t.Fatal(err)
		}
		sess = &s
		time.Sleep(expiDur / 2)
	}

	time.Sleep(expiDur / 2)

	// もう無い。
	if sess3, err := sessCont.get(sess.Id); err != nil {
		t.Fatal(err)
	} else if sess3 != nil {
		t.Error(sess3)
	}
}
