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
	"fmt"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
	"time"
)

func (sys *system) returnErrorBeforeParseRequest(w http.ResponseWriter, r *http.Request, req *session.Request, reqErr error, sess *session.Element) error {
	// リダイレクトエンドポイントが正しければ、リダイレクトでエラーを返す。

	if req == nil || req.Ta() == "" || req.RedirectUri() == "" {
		return sys.returnError(w, r, reqErr, sess)
	}

	// TA とリダイレクトエンドポイントが指定されてる。

	ta, err := sys.taDb.Get(req.Ta())
	if err != nil {
		log.Err(erro.Wrap(err))
		return sys.returnError(w, r, reqErr, sess)
	} else if ta == nil {
		log.Warn("Declared TA " + req.Ta() + " is not exist")
		return sys.returnError(w, r, reqErr, sess)
	}

	// TA は存在する。

	if !ta.RedirectUris()[req.RedirectUri()] {
		log.Warn("Declared redirect URI " + req.RedirectUri() + " is not registered")
		return sys.returnError(w, r, reqErr, sess)
	}

	// リダイレクトエンドポイントも正しい。

	sess.SetRequest(req)
	return sys.redirectError(w, r, reqErr, sess)
}

// ユーザー認証開始。
func (sys *system) authPage(w http.ResponseWriter, r *http.Request) (err error) {

	var sess *session.Element
	baseReq := newBaseRequest(r)
	if sessId := baseReq.session(); sessId != "" {
		// セッションが通知された。
		log.Debug("Session " + mosaic(sessId) + " is declared")

		if sess, err = sys.sessDb.Get(sessId); err != nil {
			log.Err(erro.Wrap(err))
			// 新規発行すれば動くので諦めない。
		} else if sess == nil {
			// セッションが無かった。
			log.Warn("Declared session " + mosaic(sessId) + " is not exist")
		} else {
			// セッションがあった。
			log.Debug("Declared session " + mosaic(sessId) + " is exist")
		}
	}

	now := time.Now()
	if sess == nil {
		// セッションを新規発行。
		sess = session.New(newId(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info("New session " + mosaic(sess.Id()) + " was generated but not yet saved")
	} else if now.After(sess.Expires().Add(-sys.sessRefDelay)) {
		// セッションを更新。
		old := sess
		sess = sess.New(newId(sys.sessLen), now.Add(sys.sessExpIn))
		log.Info("Session " + mosaic(old.Id()) + " was refreshed to " + mosaic(sess.Id()) + " but not yet saved")
	}

	// セッションは決まった。

	req, err := session.ParseRequest(r)
	if err != nil {
		return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err)), sess)
	}

	if req.Ta() == "" {
		return sys.returnError(w, r, erro.Wrap(idperr.New(idperr.Invalid_request, "no TA is declared", http.StatusBadRequest, nil)), sess)
	}

	// TA が指定されてる。
	log.Debug("TA " + req.Ta() + " is declared")

	ta, err := sys.taDb.Get(req.Ta())
	if err != nil {
		return sys.returnError(w, r, erro.Wrap(err), sess)
	} else if ta == nil {
		return sys.returnError(w, r, erro.Wrap(idperr.New(idperr.Invalid_request, "declared TA "+req.Ta()+" is not exist", http.StatusBadRequest, nil)), sess)
	}

	// TA は存在する。
	log.Debug("Declared TA " + ta.Id() + " is exist")

	// request と request_uri パラメータの読み込み。
	if req.Request() != nil {
		if req.RequestUri() != "" {
			return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(idperr.New(idperr.Invalid_request, "cannot use "+formRequest+" and "+formRequest_uri+" together", http.StatusBadRequest, nil)), sess)
		} else if keys, err := sys.keyDb.Get(); err != nil {
			return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(err), sess)
		} else if err := req.ParseRequest(req.Request(), keys, ta.Keys()); err != nil {
			return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err)), sess)
		}
	} else if req.RequestUri() != "" {
		if webElem, err := sys.webDb.Get(req.RequestUri()); err != nil {
			return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(err), sess)
		} else if keys, err := sys.keyDb.Get(); err != nil {
			return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(err), sess)
		} else if err := req.ParseRequest(webElem.Data(), keys, ta.Keys()); err != nil {
			return sys.returnErrorBeforeParseRequest(w, r, req, erro.Wrap(idperr.New(idperr.Invalid_request, erro.Unwrap(err).Error(), http.StatusBadRequest, err)), sess)
		}
	}

	if req.RedirectUri() == "" {
		return sys.returnError(w, r, erro.Wrap(idperr.New(idperr.Invalid_request, "no redirect URI is declared", http.StatusBadRequest, nil)), sess)
	} else if !ta.RedirectUris()[req.RedirectUri()] {
		return sys.returnError(w, r, erro.Wrap(idperr.New(idperr.Invalid_request, "declared redirect URI "+req.RedirectUri()+" is not registered", http.StatusBadRequest, nil)), sess)
	}

	// リダイレクトエンドポイントも正しい。
	log.Debug("Declared redirect URI " + req.RedirectUri() + " is registered")

	sess.SetRequest(req)

	// リクエストの解析は終了。
	log.Info("Authentication request " + mosaic(sess.Id()) + " reached from " + baseReq.source())

	// 重複パラメータが無いか検査。
	for k, v := range r.Form {
		if len(v) > 1 {
			return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "parameter "+k+" is overlapped", nil)), sess)
		}
	}

	if !req.Scope()[scopOpenid] {
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "scope does not have "+scopOpenid, nil)), sess)
	}

	// scope には問題無い。
	log.Debug("Declared scope has " + scopOpenid)

	switch respTypes := req.ResponseType(); len(respTypes) {
	case 0:
		return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Invalid_request, "no response type", nil)), sess)
	case 1:
		if !respTypes[respTypeCode] {
			return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Unsupported_response_type, fmt.Sprint("response type ", respTypes, " is not supported"), nil)), sess)
		}
	case 2:
		if !respTypes[respTypeCode] || !respTypes[respTypeId_token] {
			return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Unsupported_response_type, fmt.Sprint("response type ", respTypes, " is not supported"), nil)), sess)
		}
	}

	// response_type には問題無い。
	log.Debug("Response type is ", req.ResponseType())

	if req.Prompt()[prmptSelect_account] {
		if req.Prompt()[prmptNone] {
			return sys.redirectError(w, r, erro.Wrap(newErrorForRedirect(idperr.Account_selection_required, "cannot select account without UI", nil)), sess)
		}

		return sys.redirectToSelectUi(w, r, sess, "Please select your account")
	}

	return sys.afterSelect(w, r, sess, ta, nil)
}
