package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

// 認可コードと認証リクエスト時に指定されたアクセストークン用オプションの集合。
type code struct {
	// 認可コード。
	Id string `json:"id"`

	// 権利者。
	AccId string `json:"account_id"`
	// 対象 TA。
	TaId string `json:"client_id"`
	// 発行時の redirect_uri。
	RediUri string `json:"redirect_uri"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`
	// 発行するアクセストークンの有効期間。
	ExpiDur time.Duration `json:"expires_in"`

	Scops    *util.StringSet `json:"scope,omitempty"`
	Nonc     string          `json:"nonce,omitempty"`
	AuthDate time.Time       `json:"auth_time,omitempty"`
}

func newCode(codId, accId, taId, rediUri string, expiDate time.Time, expiDur time.Duration, scops map[string]bool, nonc string, authDate time.Time) *code {
	return &code{
		Id:       codId,
		AccId:    accId,
		TaId:     taId,
		RediUri:  rediUri,
		ExpiDate: expiDate,
		ExpiDur:  expiDur,
		Scops:    util.NewStringSet(scops),
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
	return this.ExpiDur
}

func (this *code) scopes() *util.StringSet {
	return this.Scops
}
