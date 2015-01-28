package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
	"strconv"
)

func responseError(w http.ResponseWriter, err error) error {
	// リダイレクトでエラーを返す時のように認証経過を廃棄する必要は無い。
	// 認証が始まって経過が記録されているなら、既にリダイレクト先が分かっているので、
	// リダイレクトでエラーを返す。

	var stat int
	m := map[string]string{}
	switch e := erro.Unwrap(err).(type) {
	case *idpError:
		log.Err(e.errorDescription())
		log.Debug(e)
		stat = e.status()
		m[formErr] = e.errorCode()
		m[formErrDesc] = e.errorDescription()
	default:
		log.Err(e)
		log.Debug(err)
		stat = http.StatusInternalServerError
		m[formErr] = errServErr
		m[formErrDesc] = e.Error()
	}

	buff, err := json.Marshal(m)
	if err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		// 最後の手段。たぶん正しい変換。
		buff = []byte(`{` + formErr + `="` + util.JsonStringEscape(m[formErr]) +
			`",` + formErrDesc + `="` + util.JsonStringEscape(m[formErrDesc]) + `"}`)
	}

	w.Header().Set("Content-Type", util.ContentTypeJson)
	w.Header().Set("Content-Length", strconv.Itoa(len(buff)))
	w.WriteHeader(stat)
	if _, err := w.Write(buff); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
	}
	return nil
}
