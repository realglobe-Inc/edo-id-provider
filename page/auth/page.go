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

// ユーザー認証ページ。
package auth

import (
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-idp-selector/request"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/rand"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

type Page struct {
	stopper *server.Stopper

	selfId string
	sigAlg string
	sigKid string

	pathSelUi  string
	pathLginUi string
	pathConsUi string
	errTmpl    *template.Template

	pwSaltLen    int
	sessLabel    string
	sessLen      int
	sessExpIn    time.Duration
	sessRefDelay time.Duration
	sessDbExpIn  time.Duration
	codLen       int
	codExpIn     time.Duration
	codDbExpIn   time.Duration
	tokExpIn     time.Duration
	jtiExpIn     time.Duration
	ticLen       int
	ticExpIn     time.Duration

	keyDb  keydb.Db
	webDb  webdb.Db
	acntDb account.Db
	consDb consent.Db
	taDb   tadb.Db
	sectDb sector.Db
	pwDb   pairwise.Db
	sessDb session.Db
	codDb  authcode.Db
	idGen  rand.Generator

	cookPath string
	cookSec  bool
	debug    bool
}

func New(
	stopper *server.Stopper,
	selfId string,
	sigAlg string,
	sigKid string,
	pathSelUi string,
	pathLginUi string,
	pathConsUi string,
	errTmpl *template.Template,
	pwSaltLen int,
	sessLabel string,
	sessLen int,
	sessExpIn time.Duration,
	sessRefDelay time.Duration,
	sessDbExpIn time.Duration,
	codLen int,
	codExpIn time.Duration,
	codDbExpIn time.Duration,
	tokExpIn time.Duration,
	jtiExpIn time.Duration,
	ticLen int,
	ticExpIn time.Duration,
	keyDb keydb.Db,
	webDb webdb.Db,
	acntDb account.Db,
	consDb consent.Db,
	taDb tadb.Db,
	sectDb sector.Db,
	pwDb pairwise.Db,
	sessDb session.Db,
	codDb authcode.Db,
	idGen rand.Generator,
	cookPath string,
	cookSec bool,
	debug bool,
) *Page {
	return &Page{
		stopper:      stopper,
		selfId:       selfId,
		sigAlg:       sigAlg,
		sigKid:       sigKid,
		pathSelUi:    pathSelUi,
		pathLginUi:   pathLginUi,
		pathConsUi:   pathConsUi,
		errTmpl:      errTmpl,
		pwSaltLen:    pwSaltLen,
		sessLabel:    sessLabel,
		sessLen:      sessLen,
		sessExpIn:    sessExpIn,
		sessRefDelay: sessRefDelay,
		sessDbExpIn:  sessDbExpIn,
		codLen:       codLen,
		codExpIn:     codExpIn,
		codDbExpIn:   codDbExpIn,
		tokExpIn:     tokExpIn,
		jtiExpIn:     jtiExpIn,
		ticLen:       ticLen,
		ticExpIn:     ticExpIn,
		keyDb:        keyDb,
		webDb:        webDb,
		acntDb:       acntDb,
		consDb:       consDb,
		taDb:         taDb,
		sectDb:       sectDb,
		pwDb:         pwDb,
		sessDb:       sessDb,
		codDb:        codDb,
		idGen:        idGen,
		cookPath:     cookPath,
		cookSec:      cookSec,
		debug:        debug,
	}
}

func (this *Page) PairwiseSaltLength() int       { return this.pwSaltLen }
func (this *Page) SectorDb() sector.Db           { return this.sectDb }
func (this *Page) PairwiseDb() pairwise.Db       { return this.pwDb }
func (this *Page) IdGenerator() rand.Generator   { return this.idGen }
func (this *Page) KeyDb() keydb.Db               { return this.keyDb }
func (this *Page) SignAlgorithm() string         { return this.sigAlg }
func (this *Page) SignKeyId() string             { return this.sigKid }
func (this *Page) SelfId() string                { return this.selfId }
func (this *Page) JwtIdExpiresIn() time.Duration { return this.jtiExpIn }

// 主にテスト用。
func (this *Page) SetSelfId(selfId string) {
	this.selfId = selfId
}

// 主にテスト用。
func (this *Page) SetCodeExpiresIn(expIn time.Duration) {
	this.codExpIn = expIn
}

func (this *Page) newCookie(sess *session.Element) *http.Cookie {
	return &http.Cookie{
		Name:     this.sessLabel,
		Value:    sess.Id(),
		Path:     this.cookPath,
		Expires:  sess.Expires(),
		Secure:   this.cookSec,
		HttpOnly: true,
	}
}

// environment のメソッドは idperr.Error を返す。
type environment struct {
	*Page

	sender *request.Request
	sess   *session.Element
}

// ユーザーエージェント向けにエラーを返す。
func (this *environment) respondErrorHtml(w http.ResponseWriter, r *http.Request, origErr error) {
	var uri *url.URL
	if this.sess.Request() != nil {
		var err error
		uri, err = url.Parse(this.sess.Request().RedirectUri())
		if err != nil {
			log.Err(this.sender, ": ", erro.Unwrap(err))
			log.Debug(this.sender, ": ", erro.Wrap(err))
		} else if this.sess.Request().State() != "" {
			q := uri.Query()
			q.Set(tagState, this.sess.Request().State())
			uri.RawQuery = q.Encode()
		}
	}

	// 経過を破棄。
	this.sess.Clear()
	if err := this.sessDb.Save(this.sess, this.sess.Expires().Add(this.sessDbExpIn-this.sessExpIn)); err != nil {
		log.Err(this.sender, ": ", erro.Wrap(err))
	} else {
		log.Debug(this.sender, ": Saved session "+logutil.Mosaic(this.sess.Id()))
	}

	if !this.sess.Saved() {
		// 未通知セッションの通知。
		http.SetCookie(w, this.newCookie(this.sess))
		log.Debug(this.sender, ": Report session "+logutil.Mosaic(this.sess.Id()))
	}

	if uri != nil {
		idperr.RedirectError(w, r, origErr, uri, this.sender)
	}

	idperr.RespondHtml(w, r, origErr, this.errTmpl, this.sender)
}

// セッション処理をしてリダイレクトさせる。
func (this *environment) redirectTo(w http.ResponseWriter, r *http.Request, uri *url.URL) {
	if err := this.sessDb.Save(this.sess, this.sess.Expires().Add(this.sessDbExpIn-this.sessExpIn)); err != nil {
		log.Err(this.sender, ": ", erro.Wrap(err))
	} else {
		log.Debug(this.sender, ": Saved session "+logutil.Mosaic(this.sess.Id()))
	}

	if !this.sess.Saved() {
		http.SetCookie(w, this.newCookie(this.sess))
		log.Debug(this.sender, ": Report session "+logutil.Mosaic(this.sess.Id()))
	}

	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)
	http.Redirect(w, r, uri.String(), http.StatusFound)
}

// リダイレクトで認可コードを返す。
func (this *environment) redirectCode(w http.ResponseWriter, r *http.Request, cod *authcode.Element, idTok string) {

	uri, err := url.Parse(this.sess.Request().RedirectUri())
	if err != nil {
		this.respondErrorHtml(w, r, erro.Wrap(err))
		return
	}

	// 経過を破棄。
	req := this.sess.Request()
	this.sess.Clear()

	q := uri.Query()
	q.Set(tagCode, cod.Id())
	if idTok != "" {
		q.Set(tagId_token, idTok)
	}
	if req.State() != "" {
		q.Set(tagState, req.State())
	}
	uri.RawQuery = q.Encode()

	log.Info(this.sender, ": Redirect "+logutil.Mosaic(this.sess.Id())+" to TA "+req.Ta())
	this.redirectTo(w, r, uri)
}
