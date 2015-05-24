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
	"github.com/realglobe-Inc/edo-id-provider/database/account"
	"github.com/realglobe-Inc/edo-id-provider/database/authcode"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-id-provider/database/coopcode"
	jtidb "github.com/realglobe-Inc/edo-id-provider/database/jti"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	"github.com/realglobe-Inc/edo-id-provider/database/pairwise"
	"github.com/realglobe-Inc/edo-id-provider/database/sector"
	"github.com/realglobe-Inc/edo-id-provider/database/session"
	"github.com/realglobe-Inc/edo-id-provider/database/token"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	webdb "github.com/realglobe-Inc/edo-idp-selector/database/web"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/go-lib/erro"
	"html/template"
	"net/http"
	"time"
)

type system struct {
	selfId  string
	sigAlg  string
	sigKid  string
	hashAlg string

	pathTok    string
	pathCoopFr string
	pathTa     string
	pathSelUi  string
	pathLginUi string
	pathConsUi string

	errTmpl *template.Template

	pwSaltLen    int
	sessLabel    string
	sessLen      int
	sessExpIn    time.Duration
	sessRefDelay time.Duration
	sessDbExpIn  time.Duration
	acodLen      int
	acodExpIn    time.Duration
	acodDbExpIn  time.Duration
	tokLen       int
	tokExpIn     time.Duration
	tokDbExpIn   time.Duration
	ccodLen      int
	ccodExpIn    time.Duration
	ccodDbExpIn  time.Duration
	jtiLen       int
	jtiExpIn     time.Duration
	jtiDbExpIn   time.Duration
	ticLen       int

	keyDb  keydb.Db
	webDb  webdb.Db
	acntDb account.Db
	consDb consent.Db
	taDb   tadb.Db
	sectDb sector.Db
	pwDb   pairwise.Db
	idpDb  idpdb.Db
	sessDb session.Db
	acodDb authcode.Db
	tokDb  token.Db
	ccodDb coopcode.Db
	jtiDb  jtidb.Db

	cookPath string
	cookSec  bool
}

func (sys *system) newCookie(sess *session.Element) *http.Cookie {
	return &http.Cookie{
		Name:     sys.sessLabel,
		Value:    sess.Id(),
		Path:     sys.cookPath,
		Expires:  sess.Expires(),
		Secure:   sys.cookSec,
		HttpOnly: true,
	}
}

// ID トークンの sub クレームとして TA に通知するアカウント ID を設定する。
func (sys *system) setSub(acnt account.Element, ta tadb.Element) error {
	if acnt.Attribute(tagSub) != nil {
		return nil
	} else if !ta.Pairwise() {
		acnt.SetAttribute(tagSub, acnt.Id())
		return nil
	}

	// セクタ固有のアカウント ID を計算。
	sect, err := sys.sectDb.Get(ta.Sector())
	if err != nil {
		return erro.Wrap(err)
	} else if sect == nil {
		sect = sector.New(ta.Sector(), randomBytes(sys.pwSaltLen))
		if existing, err := sys.sectDb.SaveIfAbsent(sect); err != nil {
			return erro.Wrap(err)
		} else if existing != nil {
			sect = existing
		}
	}
	pw := pairwise.Generate(acnt.Id(), sect.Id(), sect.Salt())

	// TA 間連携で逆引きが必要になるので、セクタ固有のアカウント ID を保存。
	if err := sys.pwDb.Save(pw); err != nil {
		return erro.Wrap(err)
	}

	acnt.SetAttribute(tagSub, pw.Pairwise())
	return nil
}

// ta に渡す acnt の ID トークンをつくる。
func (sys *system) newIdToken(ta tadb.Element, acnt account.Element, attrs map[string]bool, clms map[string]interface{}) (string, error) {
	if err := sys.setSub(acnt, ta); err != nil {
		return "", erro.Wrap(err)
	}
	keys, err := sys.keyDb.Get()
	if err != nil {
		return "", erro.Wrap(err)
	}

	now := time.Now()
	idTok := jwt.New()
	idTok.SetHeader(tagAlg, sys.sigAlg)
	if sys.sigKid != "" {
		idTok.SetHeader(tagKid, sys.sigKid)
	}
	idTok.SetClaim(tagIss, sys.selfId)
	idTok.SetClaim(tagSub, acnt.Attribute(tagSub))
	idTok.SetClaim(tagAud, ta.Id())
	idTok.SetClaim(tagExp, now.Add(sys.jtiExpIn).Unix())
	idTok.SetClaim(tagIat, now.Unix())
	for k := range attrs {
		idTok.SetClaim(k, acnt.Attribute(k))
	}
	for k, v := range clms {
		idTok.SetClaim(k, v)
	}

	if err := idTok.Sign(keys); err != nil {
		return "", erro.Wrap(err)
	}
	buff, err := idTok.Encode()
	if err != nil {
		return "", erro.Wrap(err)
	}
	return string(buff), nil
}

func (sys *system) verifyTa(ta tadb.Element, ass []byte, audSuffix string) (*jtidb.Element, error) {
	if jt, err := jwt.Parse(ass); err != nil {
		return nil, erro.Wrap(err)
	} else if jt.Header(tagAlg) == tagNone {
		return nil, erro.New("algorithm must not be " + tagNone)
	} else if err := jt.Verify(ta.Keys()); err != nil {
		return nil, erro.Wrap(err)
	} else if jt.Claim(tagIss) != ta.Id() {
		return nil, erro.New("issuer is not " + ta.Id())
	} else if jti, _ := jt.Claim(tagJti).(string); jti == "" {
		return nil, erro.New("no ID")
	} else if rawExp, _ := jt.Claim(tagExp).(float64); rawExp == 0 {
		return nil, erro.New("no expiration date")
	} else if exp := time.Unix(int64(rawExp), 0); time.Now().After(exp) {
		return nil, erro.New("expired")
	} else if jt.Claim(tagSub) != ta.Id() {
		return nil, erro.New("subject is not " + ta.Id())
	} else if aud := jt.Claim(tagAud); aud == nil {
		return nil, erro.New("no audience")
	} else if !audienceHas(aud, sys.selfId+audSuffix) {
		return nil, erro.New("audience does not contain " + sys.selfId + sys.pathTok)
	} else {
		return jtidb.New(ta.Id(), jti, exp), nil
	}
}
