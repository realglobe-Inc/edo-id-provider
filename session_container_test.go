package main

import (
	"testing"
	"time"
)

func testSessionContainer(t *testing.T, sessCont sessionContainer) {

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
	cur := time.Now()
	sess.setExpirationDate(cur.Add(testSessExpiDur))
	if err := sessCont.put(sess); err != nil {
		t.Fatal(err)
	}

	// ある。
	for end := cur.Add(2 * testSessExpiDur); cur.Before(end); cur = time.Now() {
		se, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		} else if se == nil {
			t.Fatal(cur, end)
		} else if se.id() != sess.id() {
			t.Error(cur, end, se)
		}
		se.setExpirationDate(time.Now().Add(testSessExpiDur))
		if err := sessCont.put(se); err != nil {
			t.Fatal(err)
		}
		time.Sleep(testSessExpiDur / 2)
	}

	time.Sleep(testSessExpiDur + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if se, err := sessCont.get(sess.id()); err != nil {
		t.Fatal(err)
	} else if se != nil {
		t.Error(se)
	}
}
