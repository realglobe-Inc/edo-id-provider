package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

type token struct {
	Id string `json:"id"`
	// 権利アカウント。
	AccId string `json:"account_id"`
	// 発行先 TA。
	TaId string `json:"ta_id"`

	// 有効期限。
	ExpiDate time.Time `json:"expires"`
	// リフレッシュトークン。
	RefTok string `json:"refresh_token,omitempty"`
	// 許可された scope。
	Scops util.StringSet `json:"scope,omitempty"`
	// ID トークン。
	IdTok string `json:"id_token,omitempty"`
	// 許可されたクレーム。
	Clms util.StringSet `json:"claims,omitempty"`
}

func newToken(tokId,
	accId,
	taId string,
	expiDate time.Time,
	refTok string,
	scops map[string]bool,
	idTok string,
	clms map[string]bool) *token {

	var s util.StringSet
	if len(scops) > 0 {
		s = util.NewStringSet(scops)
	}
	var c util.StringSet
	if len(clms) > 0 {
		c = util.NewStringSet(clms)
	}
	return &token{
		Id:       tokId,
		AccId:    accId,
		TaId:     taId,
		ExpiDate: expiDate,
		RefTok:   refTok,
		Scops:    s,
		IdTok:    idTok,
		Clms:     c,
	}
}

func (this *token) id() string {
	return this.Id
}

func (this *token) accountId() string {
	return this.AccId
}

func (this *token) taId() string {
	return this.TaId
}

func (this *token) expirationDate() time.Time {
	return this.ExpiDate
}

func (this *token) refreshToken() string {
	return this.RefTok
}

func (this *token) scopes() util.StringSet {
	return this.Scops
}

func (this *token) idToken() string {
	return this.IdTok
}

func (this *token) claims() util.StringSet {
	return this.Clms
}
