package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
)

func TestMongoTaKeyProvider(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg, err := NewMongoTaKeyProvider(mongoAddr, testLabel, "ta-key-provider", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if reg, err := driver.NewMongoKeyValueStore(mongoAddr, testLabel, "ta-key-provider", 0); err == nil {
			reg.Clear()
		}
	}()

	if _, err := reg.(*taKeyProvider).base.Put(testServUuid, testPublicKey); err != nil {
		t.Fatal(err)
	}

	testTaKeyProvider(t, reg)
}

func TestMongoTaKeyProviderStamp(t *testing.T) {
	if mongoAddr == "" {
		t.SkipNow()
	}

	reg, err := NewMongoTaKeyProvider(mongoAddr, testLabel, "ta-key-provider", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if reg, err := driver.NewMongoKeyValueStore(mongoAddr, testLabel, "ta-key-provider", 0); err == nil {
			reg.Clear()
		}
	}()

	if _, err := reg.(*taKeyProvider).base.Put(testServUuid, testPublicKey); err != nil {
		t.Fatal(err)
	}

	testTaKeyProviderStamp(t, reg)
}
