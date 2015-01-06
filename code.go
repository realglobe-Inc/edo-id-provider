package main

import (
	"time"
)

// 認可コードと認証リクエスト時に指定されたりしたオプションの集合。
type code struct {
	// 認可コード。
	Id string `json:"id"`
	// 発行したアカウント。
	AccId string `json:"account_id"`
	// 発行先 TA。
	TaId string `json:"ta_id"`
	// 発行時の redirect_uri。
	RediUri string `json:"redirect_uri"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`
}
