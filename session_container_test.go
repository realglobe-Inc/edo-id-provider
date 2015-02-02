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
	for end := time.Now().Add(2 * testSessExpiDur); ; {
		se, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		}
		cur := time.Now()

		// get と time.Now() の間に GC 等で時間が掛かることもあるため、
		// cur > exp でも nil が返っているとは限らない。
		// cur <= exp であれば非 nil が返らなければならない。

		if se == nil {
			if !cur.After(exp) {
				t.Fatal(cur, exp, end)
			}
			// 期限切れ。
			buff := *sess
			se = &buff
		} else if se.id() != sess.id() {
			t.Error(se, cur, exp, end)
		}

		exp = time.Now().Add(testSessExpiDur)
		se.setExpirationDate(exp)
		if err := sessCont.put(se); err != nil {
			t.Fatal(err)
		}

		if cur.After(end) {
			break
		}

		time.Sleep(exp.Sub(time.Now()) / 4)
	}

	time.Sleep(exp.Sub(time.Now()) + time.Millisecond) // redis の粒度がミリ秒のため。

	// もう無い。
	if se, err := sessCont.get(sess.id()); err != nil {
		t.Fatal(err)
	} else if se != nil {
		t.Error(se)
	}
}
