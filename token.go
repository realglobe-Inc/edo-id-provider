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
	// 有効期限。
	ExpiDate time.Time `json:"expires"`
	// リフレッシュトークン。
	RefTok string `json:"refresh_token,omitempty"`
	// scope
	Scops *util.StringSet `json:"scope,omitempty"`
}

func newToken(tokId, accId string, expiDate time.Time) *token {
	return &token{
		Id:       tokId,
		AccId:    accId,
		ExpiDate: expiDate,
	}
}

func (this *token) id() string {
	return this.Id
}

func (this *token) accountId() string {
	return this.AccId
}

func (this *token) expirationDate() time.Time {
	return this.ExpiDate
}
