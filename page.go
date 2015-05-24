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
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"net/url"
)

// ユーザーエージェント向けにエラーを返す。
func (sys *system) respondPageError(w http.ResponseWriter, r *http.Request, origErr error, sender *request.Request, sess *session.Element) (err error) {
	var uri *url.URL
	if sess.Request() != nil {
		uri, err = url.Parse(sess.Request().RedirectUri())
		if err != nil {
			log.Err(sender, ": ", erro.Unwrap(err))
			log.Debug(sender, ": ", erro.Wrap(err))
		} else if sess.Request().State() != "" {
			q := uri.Query()
			q.Set(tagState, sess.Request().State())
			uri.RawQuery = q.Encode()
		}
	}

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

	if uri != nil {
		idperr.RedirectError(w, r, origErr, uri, sender)
	}

	idperr.RespondPageError(w, r, origErr, sender, sys.errTmpl)
	return nil
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
		return sys.respondPageError(w, r, erro.Wrap(err), sender, sess)
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
