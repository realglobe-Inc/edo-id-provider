package main

import ()

type account struct {
	// IdP 内で一意かつ変更されることのない ID。
	Id string `json:"id"`
	// IdP 内で一意のログイン ID。
	Name string `json:"name"`
	// パスワード。
	Passwd string `json:"passwd"`
}
