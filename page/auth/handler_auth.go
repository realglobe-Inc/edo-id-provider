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

package auth

import (
	"fmt"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"net/http"
	"time"
)

// ユーザー認証開始。
func (this *Page) HandleAuth(w http.ResponseWriter, r *http.Request) {
	var sender *request.Request

	// panic 対策。
	defer func() {
		if rcv := recover(); rcv != nil {
			idperr.RespondHtml(w, r, erro.New(rcv), this.errTmpl, sender)
			return
		}
	}()

	if this.stopper != nil {
		this.stopper.Stop()
		defer this.stopper.Unstop()
	}

	//////////////////////////////
	server.LogRequest(level.DEBUG, r, this.debug)
	//////////////////////////////

	sender = request.Parse(r, this.sessLabel)
	log.Info(sender, ": Received authentication request")
	defer log.Info(sender, ": Handled authentication request")

	if err := this.authServe(w, r, sender); err != nil {
		idperr.RespondHtml(w, r, erro.Wrap(err), this.errTmpl, sender)
		return
	}
	return
}

func (this *Page) authServe(w http.ResponseWriter, r *http.Request, sender *request.Request) (err error) {
	var sess *session.Element
	if sessId := sender.Session(); sessId != "" {
		// セッションが通知された。
		log.Debug(sender, ": Session is declared")

		if sess, err = this.sessDb.Get(sessId); err != nil {
			log.Err(sender, ": ", erro.Wrap(err))
			// 新規発行すれば動くので諦めない。
		} else if sess == nil {
			// セッションが無かった。
			log.Warn(sender, ": Declared session is not exist")
		} else {
			// セッションがあった。
			log.Debug(sender, ": Declared session is exist")
		}
	}

	now := time.Now()
	if sess == nil {
		// セッションを新規発行。
		sess = session.New(this.idGen.String(this.sessLen), now.Add(this.sessExpIn))
		log.Info(sender, ": Generated new session "+logutil.Mosaic(sess.Id())+" but not yet saved")
	} else if now.After(sess.Expires().Add(-this.sessRefDelay)) {
		// セッションを更新。
		old := sess
		sess = sess.New(this.idGen.String(this.sessLen), now.Add(this.sessExpIn))
		log.Info(sender, ": Refreshed session "+logutil.Mosaic(old.Id())+" to "+logutil.Mosaic(sess.Id())+" but not yet saved")
	}

	// セッションは決まった。

	authReq, err := session.ParseRequest(r)
	if err != nil {
		err = erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
		return this.respondErrorHtmlBeforeGetRedirectUri(w, r, authReq, err, sender, sess)
	}

	if err := this.afterParseAuthRequest(w, r, authReq, sender, sess); err != nil {
		return this.respondErrorHtml(w, r, erro.Wrap(err), sender, sess)
	}
	return nil
}

func (this *Page) respondErrorHtmlBeforeGetRedirectUri(w http.ResponseWriter, r *http.Request, req *session.Request, origErr error, sender *request.Request, sess *session.Element) error {
	// リダイレクトエンドポイントが正しければ、リダイレクトでエラーを返す。

	if req == nil || req.Ta() == "" || req.RedirectUri() == "" {
		return this.respondErrorHtml(w, r, origErr, sender, sess)
	}

	// TA とリダイレクトエンドポイントが指定されてる。

	ta, err := this.taDb.Get(req.Ta())
	if err != nil {
		log.Err(sender, ": ", erro.Wrap(err))
		return this.respondErrorHtml(w, r, origErr, sender, sess)
	} else if ta == nil {
		log.Warn(sender, ": Declared TA "+req.Ta()+" is not exist")
		return this.respondErrorHtml(w, r, origErr, sender, sess)
	}

	// TA は存在する。
	return this.respondErrorHtmlBeforeGetRedirectUriWithTa(w, r, req, ta, origErr, sender, sess)
}

func (this *Page) respondErrorHtmlBeforeGetRedirectUriWithTa(w http.ResponseWriter, r *http.Request, req *session.Request, ta tadb.Element, origErr error, sender *request.Request, sess *session.Element) error {
	if !ta.RedirectUris()[req.RedirectUri()] {
		log.Warn(sender, ": Declared redirect URI "+req.RedirectUri()+" is not registered")
		return this.respondErrorHtml(w, r, origErr, sender, sess)
	}

	// リダイレクトエンドポイントも正しい。

	sess.SetRequest(req)
	return this.respondErrorHtml(w, r, origErr, sender, sess)
}

// request や request_uri パラメータを読み込む。
func (this *Page) parseRequestObject(req *session.Request, ta tadb.Element) error {
	if req.Request() != nil {
		if req.RequestUri() != "" {
			return erro.Wrap(idperr.New(idperr.Invalid_request, "cannot use "+tagRequest+" and "+tagRequest_uri+" together", http.StatusBadRequest, nil))
		} else if keys, err := this.keyDb.Get(); err != nil {
			return erro.Wrap(err)
		} else if err := req.ParseRequest(req.Request(), keys, ta.Keys()); err != nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
		}
	} else if req.RequestUri() != "" {
		if webElem, err := this.webDb.Get(req.RequestUri()); err != nil {
			return erro.Wrap(err)
		} else if keys, err := this.keyDb.Get(); err != nil {
			return erro.Wrap(err)
		} else if err := req.ParseRequest(webElem.Data(), keys, ta.Keys()); err != nil {
			return erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err))
		}
	}
	return nil
}

// 正しいリダイレクト URI が分かる前。
func (this *Page) afterParseAuthRequest(w http.ResponseWriter, r *http.Request, req *session.Request, sender *request.Request, sess *session.Element) error {
	ta, err := this.taDb.Get(req.Ta())
	if err != nil {
		return erro.Wrap(err)
	} else if ta == nil {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "declared TA "+req.Ta()+" is not exist", http.StatusBadRequest, nil))
	}

	// TA は存在する。
	log.Debug(sender, ": Declared TA "+ta.Id()+" is exist")

	// request と request_uri パラメータの読み込み。
	if err := this.parseRequestObject(req, ta); err != nil {
		return this.respondErrorHtmlBeforeGetRedirectUriWithTa(w, r, req, ta, erro.Wrap(err), sender, sess)
	}

	if req.RedirectUri() == "" {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "no redirect URI is declared", http.StatusBadRequest, nil))
	} else if !ta.RedirectUris()[req.RedirectUri()] {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "declared redirect URI "+req.RedirectUri()+" is not registered", http.StatusBadRequest, nil))
	}

	// リダイレクトエンドポイントも正しい。
	log.Debug(sender, ": Declared redirect URI "+req.RedirectUri()+" is registered")

	sess.SetRequest(req)
	if err := this.afterGetRedirectUri(w, r, req, ta, sender, sess); err != nil {
		return this.respondErrorHtml(w, r, erro.Wrap(err), sender, sess)
	}
	return nil
}

// 正しいリダイレクト URI が分かった後。
func (this *Page) afterGetRedirectUri(w http.ResponseWriter, r *http.Request, req *session.Request, ta tadb.Element, sender *request.Request, sess *session.Element) error {
	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "parameter "+k+" overlaps", nil))
		}
	}

	if !req.Scope()[tagOpenid] {
		return erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "scope does not have "+tagOpenid, nil))
	}

	// scope には問題無い。
	log.Debug(sender, ": Declared scope has "+tagOpenid)

	switch respTypes := req.ResponseType(); len(respTypes) {
	case 0:
		return erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "no response type", nil))
	case 1:
		if !respTypes[tagCode] {
			return erro.Wrap(newErrorForRedirect(idperr.Unsupported_response_type, fmt.Sprint("unsupported response types ", respTypes), nil))
		}
	case 2:
		if !respTypes[tagCode] || !respTypes[tagId_token] {
			return erro.Wrap(newErrorForRedirect(idperr.Unsupported_response_type, fmt.Sprint("unsupported response types ", respTypes), nil))
		}
	}

	// response_type には問題無い。
	log.Debug(sender, ": Response type is ", req.ResponseType())

	if req.Prompt()[tagSelect_account] {
		if req.Prompt()[tagNone] {
			return erro.Wrap(newErrorForRedirect(idperr.Account_selection_required, "cannot select account without UI", nil))
		}

		return this.redirectToSelectUi(w, r, sender, sess, "Please select your account")
	}

	return this.afterSelect(w, r, sender, sess, ta, nil)
}
