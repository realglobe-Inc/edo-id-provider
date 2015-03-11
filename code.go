package main

import (
	"github.com/realglobe-Inc/edo-lib/duration"
	"github.com/realglobe-Inc/edo-lib/strset"
	"time"
)

// 認可コードと認証リクエスト時に指定されたアクセストークン用オプションの集合。
type code struct {
	Id string `json:"id"`
	// 発行日時。
	Date time.Time `json:"date"`
	// 権利者アカウントの ID。
	AccId string `json:"account_id"`
	// 要求元 TA の ID。
	TaId string `json:"client_id"`
	// 発行時の redirect_uri。
	RediUri string `json:"redirect_uri"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`

	// 発行するアクセストークンの有効期間。
	ExpiDur duration.Duration `json:"expires_in"`
	// 許可された scope。
	Scops strset.StringSet `json:"scope,omitempty"`
	// 許可されたクレーム。
	Clms strset.StringSet `json:"claims,omitempty"`
	// 認証リクエストの nonce パラメータの値。
	Nonc string `json:"nonce,omitempty"`
	// 発行時の権利者アカウントの最新認証日時。
	AuthDate time.Time `json:"auth_time,omitempty"`

	// 有効か。
	Valid bool `json:"valid,omitempty"`
	// 更新日時。
	Upd time.Time `json:"update_at"`
	// 発行したアクセストークン。
	Toks strset.StringSet `json:"access_tokens,omitempty"`
}

func (this *code) copy() *code {
	c := *this
	if this.Scops != nil {
		c.Scops = strset.New(this.Scops)
	}
	if this.Clms != nil {
		c.Clms = strset.New(this.Clms)
	}
	if this.Toks != nil {
		c.Toks = strset.New(this.Toks)
	}
	return &c
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

	var s strset.StringSet
	if len(scops) > 0 {
		s = strset.New(scops)
	}
	var c strset.StringSet
	if len(clms) > 0 {
		c = strset.New(clms)
	}
	now := time.Now()
	return &code{
		Id:       codId,
		Date:     now,
		AccId:    accId,
		TaId:     taId,
		RediUri:  rediUri,
		ExpiDate: expiDate,
		ExpiDur:  duration.Duration(expiDur),
		Scops:    s,
		Clms:     c,
		Nonc:     nonc,
		AuthDate: authDate,
		Valid:    true,
		Upd:      now,
	}
}

func (this *code) id() string {
	return this.Id
}

func (this *code) date() time.Time {
	return this.Date
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

func (this *code) scopes() map[string]bool {
	return this.Scops
}

func (this *code) claims() map[string]bool {
	return this.Clms
}

func (this *code) nonce() string {
	return this.Nonc
}

func (this *code) authenticationDate() time.Time {
	return this.AuthDate
}

func (this *code) valid() bool {
	return this.Valid && !this.ExpiDate.Before(time.Now())
}

func (this *code) updateDate() time.Time {
	return this.Upd
}

func (this *code) tokens() map[string]bool {
	return this.Toks
}

// 無効にする。
func (this *code) disable() {
	this.Valid = false
	this.Upd = time.Now()
}

// 発行したアクセストークンを登録する。
func (this *code) addToken(tok string) {
	if this.Toks == nil {
		this.Toks = strset.StringSet{}
	}
	this.Toks[tok] = true
	this.Upd = time.Now()
}
