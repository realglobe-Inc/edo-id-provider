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
	"encoding/json"
	jsonutil "github.com/realglobe-Inc/edo-lib/json"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"strconv"
)

func responseError(w http.ResponseWriter, err error) {
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
		buff = []byte(`{` + formErr + `="` + jsonutil.StringEscape(m[formErr]) +
			`",` + formErrDesc + `="` + jsonutil.StringEscape(m[formErrDesc]) + `"}`)
	}

	w.Header().Set("Content-Type", server.ContentTypeJson)
	w.Header().Set("Content-Length", strconv.Itoa(len(buff)))
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	w.WriteHeader(stat)
	if _, err := w.Write(buff); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
	}
	return
}

// パニックとエラーの処理をまとめる。
func panicErrorWrapper(hndl server.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// panic時にプロセス終了しないようにrecoverする
		defer func() {
			if rcv := recover(); rcv != nil {
				responseError(w, erro.New(rcv))
				return
			}
		}()

		//////////////////////////////
		server.LogRequest(level.DEBUG, r, true)
		//////////////////////////////

		if err := hndl(w, r); err != nil {
			responseError(w, erro.Wrap(err))
			return
		}
	}
}
