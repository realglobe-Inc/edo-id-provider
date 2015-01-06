package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

type token struct {
	// アクセストークン。
	Id string `json:"id"`
	// 発行したアカウント。
	AccId string `json:"account_id"`
	// 発行先 TA。
	TaId string `json:"ta_id"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`

	// リフレッシュトークン。
	RefTok string `json:"refresh_token,omitempty"`
	// scope
	Scops *util.StringSet `json:"scope,omitempty"`
	// 許可されたクレーム。
	Clms *util.StringSet `json:"claims,omitempty"`
}

func newToken(tokId, accId, taId string, expiDate time.Time, scops map[string]bool) *token {
	return &token{
		Id:       tokId,
		AccId:    accId,
		TaId:     taId,
		ExpiDate: expiDate,

		Scops: util.NewStringSet(scops),
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

func (this *token) scopes() *util.StringSet {
	return this.Scops
}

func (this *token) claims() *util.StringSet {
	return this.Clms
}
