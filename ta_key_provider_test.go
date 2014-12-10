package main

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/realglobe-Inc/edo/driver"
	"reflect"
	"testing"
	"time"
)

// 事前に、サービス UUID a_b-c、公開鍵 testPublicKey で登録しとく。

var testPublicKey *rsa.PublicKey

func init() {
	testKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	testPublicKey = &testKey.PublicKey
}

func testTaKeyProvider(t *testing.T, reg TaKeyProvider) {
	key, _, err := reg.ServiceKey(testServUuid, nil)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(key, testPublicKey) {
		t.Error(key, testPublicKey)
	}
}

func testTaKeyProviderStamp(t *testing.T, reg TaKeyProvider) {

	key1, stmp1, err := reg.ServiceKey(testServUuid, nil)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(key1, testPublicKey) || stmp1 == nil {
		t.Error(key1, stmp1)
	}

	// キャッシュと同じだから返らない。
	key2, stmp2, err := reg.ServiceKey(testServUuid, stmp1)
	if err != nil {
		t.Fatal(err)
	} else if key2 != nil || stmp2 == nil {
		t.Error(key2, stmp2)
	}

	// キャッシュが古いから返る。
	key3, stmp3, err := reg.ServiceKey(testServUuid, &driver.Stamp{Date: stmp1.Date.Add(-time.Second), Digest: stmp1.Digest})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(key3, testPublicKey) || stmp3 == nil {
		t.Error(key3, stmp3)
	}

	// ダイジェストが違うから返る。
	key4, stmp4, err := reg.ServiceKey(testServUuid, &driver.Stamp{Date: stmp1.Date, Digest: stmp1.Digest + "a"})
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(key3, testPublicKey) || stmp4 == nil {
		t.Error(key4, stmp4)
	}
}
