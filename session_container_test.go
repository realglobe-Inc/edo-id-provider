// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// 剰余切り捨て。
func cutOff(val, thres int64) int64 {
	return val - val%thres
}

func testSessionContainer(t *testing.T, sessCont sessionContainer) {
	defer sessCont.close()

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
	bef := time.Now()
	if err := sessCont.put(sess); err != nil {
		t.Fatal(err)
	}
	diff := int64(time.Since(bef) / time.Nanosecond)

	// 期限延長テスト。
	for preSe, end := sess, time.Now().Add(2*testSessExpiDur); time.Now().Before(end); {
		bef := time.Now()
		se, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		}
		aft := time.Now()

		// get と time.Now() の間に GC 等で時間が掛かることもあるため、aft > exp でも nil が返るとは限らない。
		// だが、aft <= exp であれば非 nil が返らなければならない。
		// 同様に、bef > exp であれば nil が返らなければならない。

		if aft.UnixNano() <= cutOff(exp.UnixNano(), 1e6)-diff { // redis の粒度がミリ秒のため。
			if se == nil {
				t.Error(aft)
				t.Error(exp)
				return
			}
		} else if bef.UnixNano() > cutOff(exp.UnixNano(), 1e6)+1e6+diff { // redis の粒度がミリ秒のため。
			if se != nil {
				t.Error(bef)
				t.Error(exp)
				return
			}
			// 期限切れ。
			buff := *preSe
			se = &buff
		} else if se == nil { // bef <= exp < aft
			// 期限切れ。
			buff := *preSe
			se = &buff
		}

		if !reflect.DeepEqual(se, preSe) {
			t.Error(fmt.Sprintf("%#v", se))
			t.Error(fmt.Sprintf("%#v", preSe))
		}

		exp = time.Now().Add(testSessExpiDur)
		se.setExpirationDate(exp)
		bef = time.Now()
		if err := sessCont.put(se); err != nil {
			t.Fatal(err)
		}
		diff = int64(time.Since(bef) / time.Nanosecond)
		preSe = se

		time.Sleep(time.Millisecond)
	}

	// 消えるかどうか。
	for {
		bef := time.Now()
		se, err := sessCont.get(sess.id())
		if err != nil {
			t.Fatal(err)
		}
		aft := time.Now()

		if aft.UnixNano() <= cutOff(exp.UnixNano(), 1e6) { // redis の粒度がミリ秒のため。
			if se == nil {
				t.Error(aft)
				t.Error(exp)
				return
			}
		} else if bef.UnixNano() > cutOff(exp.UnixNano(), 1e6)+1e6 { // redis の粒度がミリ秒のため。
			if se != nil {
				t.Error(bef)
				t.Error(exp)
				return
			}
			// 消えた。
			return
		} else if se == nil { // bef <= exp < aft
			// 消えた。
			return
		}

		time.Sleep(time.Millisecond)
	}
}
