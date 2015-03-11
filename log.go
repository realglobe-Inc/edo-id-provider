package main

import (
	"github.com/realglobe-Inc/go-lib/rglog"
)

var log = rglog.Logger("github.com/realglobe-Inc/edo-id-provider")

// ログにそのまま書くのが憚られるので隠す。
func mosaic(str string) string {
	const thres = 10
	if len(str) <= thres {
		return str
	} else {
		return str[:thres] + "..."
	}
}
