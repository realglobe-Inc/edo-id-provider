package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-toolkit/util/jwt"
	"github.com/realglobe-Inc/edo-toolkit/util/strset"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ta struct {
	Id   string `json:"id"   bson:"id"`
	Name string `json:"name" bson:"name"`
	// 登録された全ての redirect_uri。
	RediUris strset.StringSet `json:"redirect_uris" bson:"redirect_uris"`
	// kid から鍵へのマップ。
	Keys keyMap `json:"keys" bson:"keys"`
	// 更新日時。
	Upd time.Time `json:"update_at" bson:"update_at"`
}

func newTa(id, name string, rediUris map[string]bool, keys map[string]interface{}) *ta {
	return &ta{
		Id:       id,
		Name:     name,
		RediUris: rediUris,
		Keys:     keys,
		Upd:      time.Now(),
	}
}

func (this *ta) id() string {
	return this.Id
}

func (this *ta) name() string {
	return this.Name
}

func (this *ta) redirectUris() map[string]bool {
	return this.RediUris
}

func (this *ta) keys() map[string]interface{} {
	return this.Keys
}

func (this *ta) updateDate() time.Time {
	return this.Upd
}

type keyMap map[string]interface{}

func (this keyMap) MarshalJSON() ([]byte, error) {
	a := []map[string]interface{}{}
	for kid, key := range this {
		m := jwt.KeyToJwkMap(key, nil)
		if kid != "" {
			m["kid"] = kid
		}
		a = append(a, m)
	}
	return json.Marshal(a)
}

func (this *keyMap) UnmarshalJSON(data []byte) error {
	var a []map[string]interface{}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	keys := map[string]interface{}{}
	for _, m := range a {
		key, err := jwt.KeyFromJwkMap(m)
		if err != nil {
			return err
		}
		kid, _ := m["kid"].(string)
		keys[kid] = key
	}
	*this = keyMap(keys)
	return nil
}

func (this keyMap) GetBSON() (interface{}, error) {
	a := []map[string]interface{}{}
	for kid, key := range this {
		m := jwt.KeyToJwkMap(key, nil)
		if kid != "" {
			m["kid"] = kid
		}
		a = append(a, m)
	}
	return a, nil
}

func (this *keyMap) SetBSON(raw bson.Raw) error {
	var a []map[string]interface{}
	if err := raw.Unmarshal(&a); err != nil {
		return err
	}
	keys := map[string]interface{}{}
	for _, m := range a {
		key, err := jwt.KeyFromJwkMap(m)
		if err != nil {
			return err
		}
		kid, _ := m["kid"].(string)
		keys[kid] = key
	}
	*this = keyMap(keys)
	return nil
}
