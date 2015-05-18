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
	"github.com/realglobe-Inc/edo-id-provider/request"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	jsonutil "github.com/realglobe-Inc/edo-lib/json"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
)

// JSON で何かを返す。
func response(w http.ResponseWriter, params map[string]interface{}) error {
	buff, err := json.Marshal(params)
	if err != nil {
		return erro.Wrap(err)
	}

	w.Header().Add("Content-Type", server.ContentTypeJson)
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	if _, err := w.Write(buff); err != nil {
		log.Err(erro.Wrap(err))
	}
	return nil
}

// エラーを返す。
func responseError(w http.ResponseWriter, origErr error, sender *request.Request) error {
	// リダイレクトでエラーを返す時のように認証経過を廃棄する必要は無い。
	// 認証が始まって経過が記録されているなら、既にリダイレクト先が
	// 分かっているので、リダイレクトでエラーを返すため。

	e := idperr.From(origErr)
	log.Err(sender, ": "+e.ErrorDescription())
	log.Debug(sender, ": ", origErr)

	buff, err := json.Marshal(map[string]string{
		tagError:             e.ErrorCode(),
		tagError_description: e.ErrorDescription(),
	})
	if err != nil {
		log.Err(sender, ": ", erro.Unwrap(err))
		log.Debug(sender, ": ", erro.Wrap(err))
		// 最後の手段。たぶん正しい変換。
		buff = []byte(`{` +
			tagError + `="` + jsonutil.StringEscape(e.ErrorCode()) + `",` +
			tagError_description + `="` + jsonutil.StringEscape(e.ErrorDescription()) +
			`"}`)
	}

	w.Header().Set("Content-Type", server.ContentTypeJson)
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	w.WriteHeader(e.Status())
	if _, err := w.Write(buff); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
	}
	return nil
}

// パニックとエラーの処理をまとめる。
func panicErrorWrapper(s *server.Stopper, f server.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Stop()
		defer s.Unstop()

		// panic時にプロセス終了しないようにrecoverする
		defer func() {
			if rcv := recover(); rcv != nil {
				responseError(w, erro.New(rcv), request.Parse(r, ""))
				return
			}
		}()

		//////////////////////////////
		server.LogRequest(level.DEBUG, r, true)
		//////////////////////////////

		if err := f(w, r); err != nil {
			responseError(w, erro.Wrap(err), request.Parse(r, ""))
			return
		}
	}
}
