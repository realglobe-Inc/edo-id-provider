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
	exp := time.Now().Add(testSessExpiDur)
	sess.setExpirationDate(exp)
	if err := sessCont.put(sess); err != nil {
		t.Fatal(err)
	}

	// 期限延長テスト。
	cur := time.Now()
	for end := cur.Add(2 * testSessExpiDur); cur.Before(end); cur = time.Now() {
		se, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		} else if cur.After(exp) {
			// 期限切れ。
			if se != nil {
				t.Fatal(se, cur, exp, end)
			}
			buff := *sess
			se = &buff
		} else if se == nil {
			t.Fatal(cur, exp, end)
		} else if se.id() != sess.id() {
			t.Error(se, cur, exp, end)
		}

		exp = time.Now().Add(testSessExpiDur)
		se.setExpirationDate(exp)
		if err := sessCont.put(se); err != nil {
			t.Fatal(err)
		}
		time.Sleep(exp.Sub(time.Now()) / 2)
	}

	time.Sleep(exp.Sub(time.Now()) + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if se, err := sessCont.get(sess.id()); err != nil {
		t.Fatal(err)
	} else if se != nil {
		t.Error(se)
	}
}
