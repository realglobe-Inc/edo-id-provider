package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

type token struct {
	Id string `json:"id"`
	// 発行日時。
	Date time.Time `json:"date"`
	// 権利者アカウントの ID。
	AccId string `json:"account_id"`
	// 要求元 TA の ID。
	TaId string `json:"ta_id"`
	// 発行時の認可コード。リフレッシュトークンと排他。
	Cod string `json:"code,omitempty"`
	// 発行時のリフレッシュトークン。認可コードと排他。
	RefTok string `json:"refresh_token,omitempty"`

	// 有効期限。
	ExpiDate time.Time `json:"expires"`
	// 許可された scope。
	Scops util.StringSet `json:"scope,omitempty"`
	// 許可されたクレーム。
	Clms util.StringSet `json:"claims,omitempty"`
	// ID トークン。
	IdTok string `json:"id_token,omitempty"`

	// 有効か。
	Valid bool `json:"valid,omitempty"`
	// 更新日時。
	Upd time.Time `json:"update_at"`
}

func newToken(tokId,
	accId,
	taId string,
	cod,
	refTok string,
	expiDate time.Time,
	scops map[string]bool,
	clms map[string]bool,
	idTok string) *token {

	now := time.Now()
	return &token{
		Id:       tokId,
		Date:     now,
		AccId:    accId,
		TaId:     taId,
		ExpiDate: expiDate,
		RefTok:   refTok,
		Scops:    scops,
		Clms:     clms,
		IdTok:    idTok,
		Valid:    true,
		Upd:      now,
	}
}

func (this *token) id() string {
	return this.Id
}

func (this *token) date() time.Time {
	return this.Date
}

func (this *token) accountId() string {
	return this.AccId
}

func (this *token) taId() string {
	return this.TaId
}

func (this *token) code() string {
	return this.Cod
}

func (this *token) refreshToken() string {
	return this.RefTok
}

func (this *token) expirationDate() time.Time {
	return this.ExpiDate
}

func (this *token) scopes() util.StringSet {
	return this.Scops
}

func (this *token) claims() util.StringSet {
	return this.Clms
}

func (this *token) idToken() string {
	return this.IdTok
}

func (this *token) valid() bool {
	return this.Valid && !this.ExpiDate.Before(time.Now())
}

func (this *token) updateDate() time.Time {
	return this.Upd
}
