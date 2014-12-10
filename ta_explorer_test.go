package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"testing"
	"time"
)

const (
	testLabel    = "edo-test"
	testServUuid = "test-service-no-uuid"
	testUri      = "http://localhost:1234/test/uri"
)

var testServExpTree *taExplorerTree

func init() {
	testServExpTree = newTaExplorerTree()
	testServExpTree.add(testUri, testServUuid)
}

func testTaExplorer(t *testing.T, reg TaExplorer) {
	servUuid, _, err := reg.ServiceUuid(testUri+"/hoge", nil)
	if err != nil {
		t.Fatal(err)
	} else if servUuid != testServUuid {
		t.Error(servUuid)
	}

	servUuid, _, err = reg.ServiceUuid(testUri, nil)
	if err != nil {
		t.Fatal(err)
	} else if servUuid != testServUuid {
		t.Error(servUuid)
	}

	servUuid, _, err = reg.ServiceUuid(testUri[:len(testUri)-1], nil)
	if err != nil {
		t.Fatal(err)
	} else if servUuid != "" {
		t.Error(servUuid)
	}
}

func testTaExplorerStamp(t *testing.T, reg TaExplorer) {

	servUuid1, stmp1, err := reg.ServiceUuid(testUri, nil)
	if err != nil {
		t.Fatal(err)
	} else if servUuid1 != testServUuid || stmp1 == nil {
		t.Error(servUuid1, stmp1)
	}

	// キャッシュと同じだから返らない。
	servUuid2, stmp2, err := reg.ServiceUuid(testUri, stmp1)
	if err != nil {
		t.Fatal(err)
	} else if servUuid2 != "" || stmp2 == nil {
		t.Error(servUuid2, stmp2)
	}

	// キャッシュが古いから返る。
	servUuid3, stmp3, err := reg.ServiceUuid(testUri, &driver.Stamp{Date: stmp1.Date.Add(-time.Second), Digest: stmp1.Digest})
	if err != nil {
		t.Fatal(err)
	} else if servUuid3 != testServUuid || stmp3 == nil {
		t.Error(servUuid3, stmp3)
	}

	// ダイジェストが違うから返る。
	servUuid4, stmp4, err := reg.ServiceUuid(testUri, &driver.Stamp{Date: stmp1.Date, Digest: stmp1.Digest + "a"})
	if err != nil {
		t.Fatal(err)
	} else if servUuid4 != testServUuid || stmp4 == nil {
		t.Error(servUuid4, stmp4)
	}
}
