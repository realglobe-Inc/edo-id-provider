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
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/edo-lib/driver"
	logutil "github.com/realglobe-Inc/edo-lib/log"
	"github.com/realglobe-Inc/edo-lib/server"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog"
	"net/http"
	"os"
	"strings"
)

func main() {
	var exitCode = 0
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()
	defer rglog.Flush()

	logutil.InitConsole("github.com/realglobe-Inc")

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	logutil.SetupConsole("github.com/realglobe-Inc", param.consLv)
	if err := logutil.Setup("github.com/realglobe-Inc", param.logType, param.logLv, param); err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	if err := serve(param); err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(erro.Wrap(err))
		exitCode = 1
		return
	}

	log.Info("Shut down")
}

func serve(param *parameters) (err error) {

	// バックエンドの準備。

	redPools := driver.NewRedisPoolSet(param.redTimeout, param.redPoolSize, param.redPoolExpIn)
	defer redPools.Close()
	monPools := driver.NewMongoPoolSet(param.monTimeout)
	defer monPools.Close()

	// 鍵。
	var keyDb keydb.Db
	switch param.keyDbType {
	case "file":
		keyDb = keydb.NewFileDb(param.keyDbPath)
		log.Info("Use keys in directory " + param.keyDbPath)
	case "redis":
		keyDb = keydb.NewRedisCache(keydb.NewFileDb(param.keyDbPath), redPools.Get(param.keyDbAddr), param.keyDbTag+"."+param.selfId, param.keyDbExpIn)
		log.Info("Use keys in directory " + param.keyDbPath + " with redis " + param.keyDbAddr + "<" + param.keyDbTag + "." + param.selfId + ">")
	default:
		return erro.New("invalid key DB type " + param.keyDbType)
	}

	// アカウント情報。
	var acntDb account.Db
	switch param.acntDbType {
	case "mongo":
		pool, err := monPools.Get(param.acntDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		acntDb = account.NewMongoDb(pool, param.acntDbTag, param.acntDbTag2)
		log.Info("Use account info in mongodb " + param.acntDbAddr + "<" + param.acntDbTag + "." + param.acntDbTag2 + ">")
	default:
		return erro.New("invalid account DB type " + param.acntDbType)
	}

	// スコープ・属性許可情報。
	var consDb consent.Db
	switch param.consDbType {
	case "memory":
		consDb = consent.NewMemoryDb()
		log.Info("Save consent info in memory")
	case "mongo":
		pool, err := monPools.Get(param.consDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		consDb = consent.NewMongoDb(pool, param.consDbTag, param.consDbTag2)
		log.Info("Save consent info in mongodb " + param.consDbAddr + "<" + param.consDbTag + "." + param.consDbTag2 + ">")
	default:
		return erro.New("invalid consent DB type " + param.consDbType)
	}

	// web データ。
	var webDb webdb.Db
	switch param.webDbType {
	case "direct":
		webDb = webdb.NewDirectDb()
		log.Info("Get web data directly")
	case "redis":
		webDb = webdb.NewRedisCache(webdb.NewDirectDb(), redPools.Get(param.webDbAddr), param.webDbTag, param.webDbExpIn)
		log.Info("Get web data with redis " + param.webDbAddr + "<" + param.webDbTag + ">")
	default:
		return erro.New("invalid web data DB type " + param.webDbType)
	}

	// TA 情報。
	var taDb tadb.Db
	switch param.taDbType {
	case "mongo":
		pool, err := monPools.Get(param.taDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		taDb = tadb.NewMongoDb(pool, param.taDbTag, param.taDbTag2, webDb)
		log.Info("Use TA info in mongodb " + param.taDbAddr + "<" + param.taDbTag + "." + param.taDbTag2 + ">")
	default:
		return erro.New("invalid TA DB type " + param.taDbType)
	}

	// セクタ固有のアカウント ID の計算に使う情報。
	var sectDb sector.Db
	switch param.sectDbType {
	case "memory":
		sectDb = sector.NewMemoryDb()
		log.Info("Save pairwise account ID calculation info in memory")
	case "mongo":
		pool, err := monPools.Get(param.sectDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		sectDb = sector.NewMongoDb(pool, param.sectDbTag, param.sectDbTag2)
		log.Info("Save pairwise account ID calculation info in mongodb " + param.sectDbAddr + "<" + param.sectDbTag + "." + param.sectDbTag2 + ">")
	default:
		return erro.New("invalid pairwise account ID calculation info DB type " + param.sectDbType)
	}

	// セクタ固有のアカウント ID 情報。
	var pwDb pairwise.Db
	switch param.pwDbType {
	case "memory":
		pwDb = pairwise.NewMemoryDb()
		log.Info("Save pairwise account IDs in memory")
	case "mongo":
		pool, err := monPools.Get(param.pwDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		pwDb = pairwise.NewMongoDb(pool, param.pwDbTag, param.pwDbTag2)
		log.Info("Save pairwise account IDs in mongodb " + param.pwDbAddr + "<" + param.pwDbTag + "." + param.pwDbTag2 + ">")
	default:
		return erro.New("invalid pairwise account ID DB type " + param.pwDbType)
	}

	// IdP 情報。
	var idpDb idpdb.Db
	switch param.idpDbType {
	case "mongo":
		pool, err := monPools.Get(param.idpDbAddr)
		if err != nil {
			return erro.Wrap(err)
		}
		idpDb = idpdb.NewMongoDb(pool, param.idpDbTag, param.idpDbTag2, webDb)
		log.Info("Use IdP info in mongodb " + param.idpDbAddr + "<" + param.idpDbTag + "." + param.idpDbTag2 + ">")
	default:
		return erro.New("invalid IdP DB type " + param.idpDbType)
	}

	// セッション。
	var sessDb session.Db
	switch param.sessDbType {
	case "memory":
		sessDb = session.NewMemoryDb()
		log.Info("Save sessions in memory")
	case "redis":
		sessDb = session.NewRedisDb(redPools.Get(param.sessDbAddr), param.sessDbTag)
		log.Info("Save sessions in redis " + param.sessDbAddr + "<" + param.sessDbTag + ">")
	default:
		return erro.New("invalid session DB type " + param.sessDbType)
	}

	// 認可コード。
	var acodDb authcode.Db
	switch param.acodDbType {
	case "memory":
		acodDb = authcode.NewMemoryDb()
		log.Info("Save authorization codes in memory")
	case "redis":
		acodDb = authcode.NewRedisDb(redPools.Get(param.acodDbAddr), param.acodDbTag)
		log.Info("Save authorization codes in redis " + param.acodDbAddr + "<" + param.acodDbTag + ">")
	default:
		return erro.New("invalid authorization code DB type " + param.acodDbType)
	}

	// アクセストークン。
	var tokDb token.Db
	switch param.tokDbType {
	case "memory":
		tokDb = token.NewMemoryDb()
		log.Info("Save access tokens in memory")
	case "redis":
		tokDb = token.NewRedisDb(redPools.Get(param.tokDbAddr), param.tokDbTag)
		log.Info("Save access tokens in redis " + param.tokDbAddr + "<" + param.tokDbTag + ">")
	default:
		return erro.New("invalid access token DB type " + param.tokDbType)
	}

	// 仲介コード。
	var ccodDb coopcode.Db
	switch param.ccodDbType {
	case "memory":
		ccodDb = coopcode.NewMemoryDb()
		log.Info("Save cooperation codes in memory")
	case "redis":
		ccodDb = coopcode.NewRedisDb(redPools.Get(param.ccodDbAddr), param.ccodDbTag)
		log.Info("Save cooperation codes in redis " + param.ccodDbAddr + "<" + param.ccodDbTag + ">")
	default:
		return erro.New("invalid cooperation code DB type " + param.ccodDbType)
	}

	// JWT の ID。
	var jtiDb jtidb.Db
	switch param.jtiDbType {
	case "memory":
		jtiDb = jtidb.NewMemoryDb()
		log.Info("Save JWT IDs in memory")
	case "redis":
		jtiDb = jtidb.NewRedisDb(redPools.Get(param.jtiDbAddr), param.jtiDbTag)
		log.Info("Save JWT IDs in redis " + param.jtiDbAddr + "<" + param.jtiDbTag + ">")
	default:
		return erro.New("invalid JWT ID DB type " + param.jtiDbType)
	}

	sys := &system{
		param.selfId,
		param.sigAlg,
		param.sigKid,

		param.pathTok,
		param.pathTa,
		param.pathSelUi,
		param.pathLginUi,
		param.pathConsUi,
		param.pathErrUi,

		param.pwSaltLen,
		param.sessLabel,
		param.sessLen,
		param.sessExpIn,
		param.sessRefDelay,
		param.sessDbExpIn,
		param.acodLen,
		param.acodExpIn,
		param.acodDbExpIn,
		param.tokLen,
		param.tokExpIn,
		param.tokDbExpIn,
		param.ccodLen,
		param.ccodExpIn,
		param.ccodDbExpIn,
		param.jtiLen,
		param.jtiExpIn,
		param.jtiDbExpIn,
		param.ticLen,

		keyDb,
		webDb,
		acntDb,
		consDb,
		taDb,
		sectDb,
		pwDb,
		idpDb,
		sessDb,
		acodDb,
		tokDb,
		ccodDb,
		jtiDb,

		param.cookPath,
		param.cookSec,
	}

	// バックエンドの準備完了。

	s := server.NewStopper()
	defer func() {
		// 処理の終了待ち。
		s.Lock()
		defer s.Unlock()
		for s.Stopped() {
			s.Wait()
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/", panicErrorWrapper(s, func(w http.ResponseWriter, r *http.Request) error {
		return erro.Wrap(idperr.New(idperr.Invalid_request, "invalid endpoint", http.StatusNotFound, nil))
	}))
	mux.HandleFunc(param.pathOk, panicErrorWrapper(s, func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}))
	mux.HandleFunc(param.pathAuth, panicErrorWrapper(s, sys.authPage))
	mux.HandleFunc(param.pathSel, panicErrorWrapper(s, sys.selectPage))
	mux.HandleFunc(param.pathLgin, panicErrorWrapper(s, sys.lginPage))
	mux.HandleFunc(param.pathCons, panicErrorWrapper(s, sys.consentPage))
	mux.HandleFunc(param.pathTa, panicErrorWrapper(s, sys.taApiHandler().ServeHTTP))
	mux.HandleFunc(param.pathTok, panicErrorWrapper(s, sys.tokenApi))
	mux.HandleFunc(param.pathAcnt, panicErrorWrapper(s, sys.accountApi))
	mux.HandleFunc(param.pathCoopFr, panicErrorWrapper(s, sys.cooperateFromApi))
	mux.HandleFunc(param.pathCoopTo, panicErrorWrapper(s, sys.cooperateToApi))
	if param.uiDir != "" {
		// ファイル配信も自前でやる。
		pathUi := strings.TrimRight(param.pathUi, "/") + "/"
		mux.Handle(pathUi, http.StripPrefix(pathUi, http.FileServer(http.Dir(param.uiDir))))
	}

	return server.Serve(param, mux)
}
