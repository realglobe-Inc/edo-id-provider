package main

import (
	"net/http"
	"net/url"
)

type request interface {
	// 元になる *http.Request を返す。
	raw() *http.Request
}

type errorRedirectRequest interface {
	request
	// 処理後に飛ばすリダイレクト先 URI を返す。
	redirectUri() *url.URL
}
