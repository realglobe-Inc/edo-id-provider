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
	s := r.FormValue(key)
	if s == "" {
		return map[string]bool{}
	}
	return strset.FromSlice(strings.Split(s, " "))
}

// フォーム値用にスペース区切りにして返す。
func valueSetToForm(m map[string]bool) string {
	buff := ""
	for v, ok := range m {
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

// スペース区切りのフォーム値を配列にして返す。
func formValues(r *http.Request, key string) []string {
	s := r.FormValue(key)
	if s == "" {
		return []string{}
	}
	return strings.Split(s, " ")
}

// フォーム値用にスペース区切りにして返す。
func valuesToForm(s []string) string {
	buff := ""
	for _, v := range s {
		if v == "" {
			continue
		}

		if len(buff) > 0 {
			buff += " "
		}
		buff += v
	}
	return buff
}
