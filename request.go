package main

import (
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"net/url"
	"strings"
)

type errorRedirectRequest interface {
	// 処理後に飛ばすリダイレクト先 URI を返す。
	redirectUri() *url.URL
}

// ブラウザからのリクエスト。
type browserRequest struct {
	sess string
}

func newBrowserRequest(r *http.Request) *browserRequest {
	var sess string
	if cook, err := r.Cookie(cookSess); err != nil {
		if err != http.ErrNoCookie {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
		}
	} else {
		sess = cook.Value
	}
	return &browserRequest{sess: sess}
}

func (this *browserRequest) session() string {
	return this.sess
}

// スペース区切りのフォーム値を集合にして返す。
func formValueSet(r *http.Request, key string) map[string]bool {
	s := r.FormValue(key)
	set := map[string]bool{}
	for _, v := range strings.Split(s, " ") {
		set[v] = true
	}
	return set
}

// フォーム値用にスペース区切りにして返す。
func valueSetToForm(v map[string]bool) string {
	buff := ""
	for v, ok := range v {
		if !ok || v == "" {
			continue
		}

		if len(buff) > 0 {
			buff += " "
		}
		buff += v
	}
	return buff
}
