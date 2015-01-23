package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/util"
)

type ta struct {
	_id   string
	_name string
	// 登録された全ての redirect_uri。
	_rediUris util.StringSet
	// kid から公開鍵へのマップ。
	_keys map[string]crypto.PublicKey
}

func newTa(id, name string, rediUris map[string]bool, keys map[string]crypto.PublicKey) *ta {
	return &ta{
		_id:       id,
		_name:     name,
		_rediUris: rediUris,
		_keys:     keys,
	}
}
func (this *ta) id() string {
	return this._id
}

func (this *ta) name() string {
	return this._name
}

func (this *ta) redirectUris() map[string]bool {
	return this._rediUris
}

func (this *ta) keys() map[string]crypto.PublicKey {
	return this._keys
}
