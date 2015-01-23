package main

import (
	"reflect"
	"testing"
)

func testConsentContainer(t *testing.T, consCont consentContainer) {
	// まだ無い。
	if scops, clms, err := consCont.get(testAcc.id(), testTa.id); err != nil {
		t.Fatal(err)
	} else if scops != nil {
		t.Error(scops)
	} else if clms != nil {
		t.Error(clms)
	}

	testScops := map[string]bool{"openid": true, "profile": true}
	testClms := map[string]bool{"website": true, "email": true}
	// 入れる。
	if err := consCont.put(testAcc.id(), testTa.id, testScops, testClms, nil, nil); err != nil {
		t.Fatal(err)
	}

	// ある。
	if scops, clms, err := consCont.get(testAcc.id(), testTa.id); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(scops, testScops) {
		t.Error(scops, testScops)
	} else if !reflect.DeepEqual(clms, testClms) {
		t.Error(clms, testClms)
	}

	// scope から profile を、claim から email を削除。
	if err := consCont.put(testAcc.id(), testTa.id, nil, nil, map[string]bool{"profile": true}, map[string]bool{"email": true}); err != nil {
		t.Fatal(err)
	}

	// 変更が適用されてる。
	if scops, clms, err := consCont.get(testAcc.id(), testTa.id); err != nil {
		t.Fatal(err)
	} else if s := map[string]bool{"openid": true}; !reflect.DeepEqual(scops, s) {
		t.Error(scops, s)
	} else if c := map[string]bool{"website": true}; !reflect.DeepEqual(clms, c) {
		t.Error(clms, c)
	}

	// scope に profile を、claim に email を追加。
	if err := consCont.put(testAcc.id(), testTa.id, map[string]bool{"profile": true}, map[string]bool{"email": true}, nil, nil); err != nil {
		t.Fatal(err)
	}

	// 変更が適用されてる。
	if scops, clms, err := consCont.get(testAcc.id(), testTa.id); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(scops, testScops) {
		t.Error(scops, testScops)
	} else if !reflect.DeepEqual(clms, testClms) {
		t.Error(clms, testClms)
	}
}
