package main

import (
	"crypto"
	"encoding/json"
	"github.com/realglobe-Inc/edo/util/jwt"
	"github.com/realglobe-Inc/edo/util/strset"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ta struct {
	Id   string `json:"id"   bson:"id"`
	Name string `json:"name" bson:"name"`
	// 登録された全ての redirect_uri。
	RediUris strset.StringSet `json:"redirect_uris" bson:"redirect_uris"`
	// kid から公開鍵へのマップ。
	Keys publicKeyMap `json:"keys" bson:"keys"`
	// 更新日時。
	Upd time.Time `json:"update_at" bson:"update_at"`
}

func newTa(id, name string, rediUris map[string]bool, keys map[string]crypto.PublicKey) *ta {
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

func (this *ta) keys() map[string]crypto.PublicKey {
	return this.Keys
}

func (this *ta) updateDate() time.Time {
	return this.Upd
}

type publicKeyMap map[string]crypto.PublicKey

func (this publicKeyMap) MarshalJSON() ([]byte, error) {
	a := []map[string]interface{}{}
	for kid, key := range this {
		m := jwt.PublicKeyToJwkMap(kid, key)
		a = append(a, m)
	}
	return json.Marshal(a)
}

func (this *publicKeyMap) UnmarshalJSON(data []byte) error {
	var a []map[string]interface{}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	keys := map[string]crypto.PublicKey{}
	for _, m := range a {
		kid, key, err := jwt.PublicKeyFromJwkMap(m)
		if err != nil {
			return err
		}
		keys[kid] = key
	}
	*this = publicKeyMap(keys)
	return nil
}

func (this publicKeyMap) GetBSON() (interface{}, error) {
	a := []map[string]interface{}{}
	for kid, key := range this {
		m := jwt.PublicKeyToJwkMap(kid, key)
		a = append(a, m)
	}
	return a, nil
}

func (this *publicKeyMap) SetBSON(raw bson.Raw) error {
	var a []map[string]interface{}
	if err := raw.Unmarshal(&a); err != nil {
		return err
	}
	keys := map[string]crypto.PublicKey{}
	for _, m := range a {
		kid, key, err := jwt.PublicKeyFromJwkMap(m)
		if err != nil {
			return err
		}
		keys[kid] = key
	}
	*this = publicKeyMap(keys)
	return nil
}
