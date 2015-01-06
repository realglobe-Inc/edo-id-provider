package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
)

type account struct {
	m map[string]interface{}

	// IdP 内で一意かつ変更されることのない ID。
	accId string
	// IdP 内で一意のログイン ID。
	accName string
	// パスワード。
	passwd string
}

func newAccount(m map[string]interface{}) *account {
	a := &account{m: map[string]interface{}{}}
	for k, v := range m {
		a.m[k] = v
	}
	return a
}

func (this *account) id() string {
	if this.accId == "" {
		this.accId, _ = this.m["id"].(string)
	}
	return this.accId
}

func (this *account) name() string {
	if this.accName == "" {
		this.accName, _ = this.m["username"].(string)
	}
	return this.accName
}

func (this *account) password() string {
	if this.passwd == "" {
		this.passwd, _ = this.m["password"].(string)
	}
	return this.passwd
}

func (this *account) attribute(attrName string) interface{} {
	return this.m[attrName]
}

func (this *account) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.m)
}

func (this *account) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &this.m)
}

func (this *account) GetBSON() (interface{}, error) {
	return this.m, nil
}

func (this *account) SetBSON(raw bson.Raw) error {
	return raw.Unmarshal(&this.m)
}
