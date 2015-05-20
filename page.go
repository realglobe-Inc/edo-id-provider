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
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// リダイレクトでエラーを返す。
func (sys *system) redirectErrorTo(w http.ResponseWriter, r *http.Request, origErr error, uri *url.URL, queries map[string]string, sender *request.Request, sess *session.Element) error {
	// 経過を破棄。
	sess.Clear()
	if err := sys.sessDb.Save(sess, sess.Expires().Add(sys.sessDbExpIn-sys.sessExpIn)); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
	} else {
		log.Debug(sender, ": Saved session "+mosaic(sess.Id()))
	}

	if !sess.Saved() {
		http.SetCookie(w, sys.newCookie(sess))
		log.Debug(sender, ": Report session "+mosaic(sess.Id()))
	}

	// エラー内容の添付。
	e := idperr.From(origErr)
	log.Err(sender, ": "+e.ErrorDescription())
	log.Debug(sender, ": ", origErr)

	q := uri.Query()
	q.Set(tagError, e.ErrorCode())
	q.Set(tagError_description, e.ErrorDescription())
	for k, v := range queries {
		q.Set(k, v)
	}
	uri.RawQuery = q.Encode()

	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)
	http.Redirect(w, r, uri.String(), http.StatusFound)
	return nil
}

// ユーザーエージェント向けにエラーを返す。
func (sys *system) returnError(w http.ResponseWriter, r *http.Request, origErr error, sender *request.Request, sess *session.Element) error {

	if sys.pathErrUi != "" {
		uri, err := url.Parse(sys.pathErrUi)
		if err == nil {
			return sys.redirectErrorTo(w, r, origErr, uri, nil, sender, sess)
		}
		log.Err(sender, ": ", erro.Wrap(err))
	}

	// 自前でユーザー向けの HTML を返さなきゃならない。

	// 経過を破棄。
	sess.Clear()
	if err := sys.sessDb.Save(sess, sess.Expires().Add(sys.sessDbExpIn-sys.sessExpIn)); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
	} else {
		log.Debug(sender, ": Saved session "+mosaic(sess.Id()))
	}

	if !sess.Saved() {
		// 未通知セッションの通知。
		http.SetCookie(w, sys.newCookie(sess))
		log.Debug(sender, ": Report session "+mosaic(sess.Id()))
	}

	e := idperr.From(origErr)
	log.Err(sender, ": "+e.ErrorDescription())
	log.Debug(sender, ": ", origErr)

	msg := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Error</title></head><body><h1>`
	msg += strconv.Itoa(e.Status())
	msg += " "
	msg += http.StatusText(e.Status())
	msg += `</h1><p><font size="+1"><b>`
	msg += strings.Replace(html.EscapeString(e.ErrorCode()), "\n", "<br/>", -1)
	msg += " "
	msg += strings.Replace(html.EscapeString(e.ErrorDescription()), "\n", "<br/>", -1)
	msg += `:</b></font></p><p>`
	msg += html.EscapeString(e.Error())
	msg += `</p></body></html>`
	buff := []byte(msg)

	w.Header().Set(tagContent_type, server.ContentTypeHtml)
	w.Header().Set(tagContent_length, strconv.Itoa(len(buff)))
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)
	w.WriteHeader(e.Status())
	if _, err := w.Write(buff); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
	}
	return nil
}

// TA 向けにリダイレクトでエラーを返す。
func (sys *system) redirectError(w http.ResponseWriter, r *http.Request, origErr error, sender *request.Request, sess *session.Element) error {
	uri, err := url.Parse(sess.Request().RedirectUri())
	if err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
		return sys.returnError(w, r, origErr, sender, sess)
	}

	var queries map[string]string
	if stat := sess.Request().State(); stat != "" {
		queries = map[string]string{tagState: stat}
	}
	return sys.redirectErrorTo(w, r, origErr, uri, queries, sender, sess)
}

// セッション処理をしてリダイレクトさせる。
func (sys *system) redirectTo(w http.ResponseWriter, r *http.Request, uri *url.URL, sender *request.Request, sess *session.Element) error {
	if err := sys.sessDb.Save(sess, sess.Expires().Add(sys.sessDbExpIn-sys.sessExpIn)); err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
	} else {
		log.Debug(sender, ": Saved session "+mosaic(sess.Id()))
	}

	if !sess.Saved() {
		http.SetCookie(w, sys.newCookie(sess))
		log.Debug(sender, ": Report session "+mosaic(sess.Id()))
	}

	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)
	http.Redirect(w, r, uri.String(), http.StatusFound)
	return nil
}

// リダイレクトで認可コードを返す。
func (sys *system) redirectCode(w http.ResponseWriter, r *http.Request, cod *authcode.Element, idTok string, sender *request.Request, sess *session.Element) error {

	uri, err := url.Parse(sess.Request().RedirectUri())
	if err != nil {
		return sys.returnError(w, r, erro.Wrap(err), sender, sess)
	}

	// 経過を破棄。
	req := sess.Request()
	sess.Clear()

	q := uri.Query()
	q.Set(tagCode, cod.Id())
	if idTok != "" {
		q.Set(tagId_token, idTok)
	}
	if req.State() != "" {
		q.Set(tagState, req.State())
	}
	uri.RawQuery = q.Encode()

	log.Info(sender, ": Redirect "+mosaic(sess.Id())+" to TA "+req.Ta())
	return sys.redirectTo(w, r, uri, sender, sess)
}
