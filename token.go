package main

import (
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
	Scops map[string]bool `json:"scope,omitempty"`
}
