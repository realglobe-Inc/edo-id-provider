package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

// 認可コードと認証リクエスト時に指定されたアクセストークン用オプションの集合。
type code struct {
	// 認可コード。
	Id string `json:"id"`
	// 権利アカウント。
	AccId string `json:"account_id"`
	// 対象 TA。
	TaId string `json:"client_id"`
	// 発行時の redirect_uri。
	RediUri string `json:"redirect_uri"`

	// 認可コードの有効期限。
	ExpiDate time.Time `json:"expires"`
	// 発行するアクセストークンの有効期間。
	ExpiDur util.Duration `json:"expires_in"`
	// 許可された scope。
	Scops util.StringSet `json:"scope,omitempty"`
	// 許可されたクレーム。
	Clms util.StringSet `json:"claims,omitempty"`
	// 認証リクエストの nonce パラメータの値。
	Nonc string `json:"nonce,omitempty"`
	// 権利アカウントの最新認証日時。
	AuthDate time.Time `json:"auth_time,omitempty"`
}

func newCode(codId,
	accId,
	taId,
	rediUri string,
	expiDate time.Time,
	expiDur time.Duration,
	scops,
	clms map[string]bool,
	nonc string,
	authDate time.Time) *code {

	var s util.StringSet
	if len(scops) > 0 {
		s = util.NewStringSet(scops)
	}
	var c util.StringSet
	if len(clms) > 0 {
		c = util.NewStringSet(clms)
	}
	return &code{
		Id:       codId,
		AccId:    accId,
		TaId:     taId,
		RediUri:  rediUri,
		ExpiDate: expiDate,
		ExpiDur:  util.Duration(expiDur),
		Scops:    s,
		Clms:     c,
		Nonc:     nonc,
		AuthDate: authDate,
	}
}

func (this *code) id() string {
	return this.Id
}

func (this *code) accountId() string {
	return this.AccId
}

func (this *code) taId() string {
	return this.TaId
}

func (this *code) redirectUri() string {
	return this.RediUri
}

func (this *code) expirationDate() time.Time {
	return this.ExpiDate
}

func (this *code) expirationDuration() time.Duration {
	return time.Duration(this.ExpiDur)
}

func (this *code) scopes() util.StringSet {
	return this.Scops
}

func (this *code) claims() util.StringSet {
	return this.Clms
}

func (this *code) nonce() string {
	return this.Nonc
}

func (this *code) authenticationDate() time.Time {
	return this.AuthDate
}
