package main

import (
	"github.com/realglobe-Inc/edo/util/strset"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"strings"
)

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
	set := map[string]bool{}
	s := r.FormValue(key)
	if s == "" {
		return set
	}
	return strset.FromSlice(strings.Split(s, " "))
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
