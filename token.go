// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/realglobe-Inc/edo-lib/strset"
	"time"
)

type token struct {
	Id string `json:"id"`
	// 発行日時。
	Date time.Time `json:"date"`
	// 権利者アカウントの ID。
	AccId string `json:"account_id"`
	// 要求元 TA の ID。
	TaId string `json:"client_id"`
	// 発行時の認可コード。リフレッシュトークンと排他。
	Cod string `json:"code,omitempty"`
	// 発行時のリフレッシュトークン。認可コードと排他。
	RefTok string `json:"refresh_token,omitempty"`

	// 有効期限。
	ExpiDate time.Time `json:"expires"`
	// 許可された scope。
	Scops strset.StringSet `json:"scope,omitempty"`
	// 許可されたクレーム。
	Clms strset.StringSet `json:"claims,omitempty"`
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

func (this *token) scopes() map[string]bool {
	return this.Scops
}

func (this *token) claims() map[string]bool {
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

func (this *token) disable() {
	this.Valid = false
}
