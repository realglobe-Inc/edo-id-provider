package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"reflect"
	"testing"
)

const (
	testUsrUuid  = "test-user-no-uuid"
	testAttrName = "test-attribute-no-name"
)

var testAttr = map[string]interface{}{"array": []interface{}{"elem-1", "elem-2"}}

// JSON を通して等しいかどうか調べる。
func jsonEqual(v1 interface{}, v2 interface{}) (equal bool) {
	b1, err := json.Marshal(v1)
	if err != nil {
		log.Err(erro.Wrap(err))
		return false
	}
	var w1 interface{}
	if err := json.Unmarshal(b1, &w1); err != nil {
		log.Err(erro.Wrap(err))
		return false
	}

	b2, err := json.Marshal(v2)
	if err != nil {
		log.Err(erro.Wrap(err))
		return false
	}
	var w2 interface{}
	if err := json.Unmarshal(b2, &w2); err != nil {
		log.Err(erro.Wrap(err))
		return false
	}

	return reflect.DeepEqual(w1, w2)
}

// 要事前登録。

func testUserAttributeRegistry(t *testing.T, reg UserAttributeRegistry) {
	usrUuid := testUsrUuid
	attrName := testAttrName
	usrAttr := testAttr

	usrAttr1, _, err := reg.UserAttribute(usrUuid, attrName, nil)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(usrAttr1, usrAttr) {
		if !jsonEqual(usrAttr1, usrAttr) {
			t.Error(usrAttr1)
		}
	}

	usrAttr2, _, err := reg.UserAttribute(usrUuid, attrName+"1", nil)
	if err != nil {
		t.Fatal(err)
	} else if usrAttr2 != nil {
		t.Error(usrAttr2)
	}

	usrAttr3, _, err := reg.UserAttribute(usrUuid+"1", attrName, nil)
	if err != nil {
		t.Fatal(err)
	} else if usrAttr3 != nil {
		t.Error(usrAttr3)
	}
}
