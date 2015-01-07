package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/util"
)

type ta struct {
	id   string
	name string
	// 登録された全ての redirect_uri。
	rediUris util.StringSet
	// kid から公開鍵へのマップ。
	pubKeys map[string]crypto.PublicKey
}

func (this *ta) hasRedirectUri(rediUri string) bool {
	return this.rediUris[rediUri]
}

func (this *ta) keys() map[string]crypto.PublicKey {
	return this.pubKeys
}
