package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
)

func TestMongoTaKeyProvider(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg := NewMongoTaKeyProvider(mongoAddr, testLabel, "ta-key-provider", 0, 0)
	defer driver.NewMongoKeyValueStore(mongoAddr, testLabel, "ta-key-provider", nil, nil, nil, 0, 0).Clear()

	if _, err := reg.(*taKeyProvider).base.Put(testServUuid, testPublicKey); err != nil {
		t.Fatal(err)
	}

	testTaKeyProvider(t, reg)
}

func TestMongoTaKeyProviderStamp(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg := NewMongoTaKeyProvider(mongoAddr, testLabel, "ta-key-provider", 0, 0)
	defer driver.NewMongoKeyValueStore(mongoAddr, testLabel, "ta-key-provider", nil, nil, nil, 0, 0).Clear()

	if _, err := reg.(*taKeyProvider).base.Put(testServUuid, testPublicKey); err != nil {
		t.Fatal(err)
	}

	testTaKeyProviderStamp(t, reg)
}
