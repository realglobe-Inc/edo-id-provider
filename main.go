package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"github.com/realglobe-Inc/go-lib-rg/rglog"
	"net/http"
	"os"
	"time"
)

var exitCode = 0

func exit() {
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func main() {
	defer exit()
	defer rglog.Flush()

	util.InitConsoleLog("github.com/realglobe-Inc")

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	util.SetupConsoleLog("github.com/realglobe-Inc", param.consLv)
	if err := util.SetupLog("github.com/realglobe-Inc", param.logType, param.logLv, param.logPath, param.fluAddr, param.fluTag); err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	if err := mainCore(param); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	log.Info("Shut down.")
}

// system を準備する。
func mainCore(param *parameters) error {
	const (
		connNum = 5
		idlDur  = 10 * time.Minute
	)
	redPools := map[string]*redis.Pool{} // 同じ redis-server ならコネクションプールを共有する。

	var taCont taContainer
	switch param.taContType {
	case "file":
		taCont = newFileTaContainer(param.taContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file TA container " + param.taContPath)
	case "mongo":
		taCont = newMongoTaContainer(param.taContUrl, param.taContDb, param.taContColl, param.caStaleDur, param.caExpiDur)
		log.Info("Use mongodb TA container " + param.taContUrl)
	default:
		return erro.New("invalid TA container type " + param.taContType)
	}

	var accCont accountContainer
	switch param.accContType {
	case "file":
		accCont = newFileAccountContainer(param.accContPath, param.accNameContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file account container " + param.accContPath + "," + param.accNameContPath)
	case "mongo":
		accCont = newMongoAccountContainer(param.accContUrl, param.accContDb, param.accContColl, param.caStaleDur, param.caExpiDur)
		log.Info("Use mongodb account container " + param.accContUrl)
	default:
		return erro.New("invalid account container type " + param.accContType)
	}

	var consCont consentContainer
	switch param.consContType {
	case "file":
		consCont = newFileConsentContainer(param.consContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file consent container " + param.consContPath)
	case "mongo":
		consCont = newMongoConsentContainer(param.consContUrl, param.consContDb, param.consContColl, param.caStaleDur, param.caExpiDur)
		log.Info("Use mongodb consent container " + param.consContUrl)
	default:
		return erro.New("invalid consent container type " + param.consContType)
	}

	var sessCont sessionContainer
	switch param.sessContType {
	case "memory":
		sessCont = newMemorySessionContainer(param.sessIdLen, param.procId, param.caStaleDur, param.caExpiDur)
		log.Info("Use memory session container.")
	case "file":
		sessCont = newFileSessionContainer(param.sessIdLen, param.procId, param.sessContPath, param.sessExpiContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file session container " + param.sessContPath + "," + param.sessExpiContPath)
	case "redis":
		if redPools[param.sessContUrl] == nil {
			redPools[param.sessContUrl] = driver.NewRedisPool(param.sessContUrl, connNum, idlDur)
		}
		sessCont = newRedisSessionContainer(param.sessIdLen, param.procId, redPools[param.sessContUrl], param.sessContPrefix, param.caStaleDur, param.caExpiDur)
		log.Info("Use redis session container " + param.sessContUrl)
	default:
		return erro.New("invalid session container type " + param.sessContType)
	}

	var codCont codeContainer
	switch param.codContType {
	case "memory":
		codCont = newMemoryCodeContainer(param.codIdLen, param.procId, param.codTicExpDur, param.codSavDur, param.caStaleDur, param.caExpiDur)
		log.Info("Use memory code container.")
	case "file":
		codCont = newFileCodeContainer(param.codIdLen, param.procId, param.codTicExpDur, param.codSavDur, param.codContPath, param.codExpiContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file code container " + param.codContPath + "," + param.codExpiContPath)
	case "redis":
		if redPools[param.codContUrl] == nil {
			redPools[param.codContUrl] = driver.NewRedisPool(param.codContUrl, connNum, idlDur)
		}
		codCont = newRedisCodeContainer(param.codIdLen, param.procId, param.codTicExpDur, param.codSavDur, redPools[param.codContUrl], param.codContPrefix, param.caStaleDur, param.caExpiDur)
		log.Info("Use redis code container " + param.codContUrl)
	default:
		return erro.New("invalid code container type " + param.codContType)
	}

	key, err := util.ReadPrivateKey(param.keyPath)
	if err != nil {
		return erro.Wrap(err)
	}

	var tokCont tokenContainer
	switch param.tokContType {
	case "memory":
		tokCont = newMemoryTokenContainer(param.tokIdLen, param.procId, param.tokSavDur, param.caStaleDur, param.caExpiDur)
		log.Info("Use memory token container.")
	case "file":
		tokCont = newFileTokenContainer(param.tokIdLen, param.procId, param.tokSavDur, param.tokContPath, param.tokExpiContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file token container " + param.tokContPath + "," + param.tokExpiContPath)
	case "redis":
		if redPools[param.tokContUrl] == nil {
			redPools[param.tokContUrl] = driver.NewRedisPool(param.tokContUrl, connNum, idlDur)
		}
		tokCont = newRedisTokenContainer(param.tokIdLen, param.procId, param.tokSavDur, redPools[param.tokContUrl], param.tokContPrefix, param.caStaleDur, param.caExpiDur)
		log.Info("Use redis token container " + param.tokContUrl)
	default:
		return erro.New("invalid token container type " + param.tokContType)
	}

	sys := &system{
		param.selfId,
		param.secCook,
		param.codIdLen / 2,
		param.codIdLen / 2,
		param.uiUri,
		param.uiPath,
		taCont,
		accCont,
		consCont,
		sessCont,
		codCont,
		tokCont,
		param.codExpiDur,
		param.tokExpiDur,
		param.idTokExpiDur,
		param.sessExpiDur,
		param.sigAlg,
		param.kid,
		key,
	}
	return serve(sys, param.socType, param.socPath, param.socPort, param.protType, nil)
}

// 振り分ける。
func serve(sys *system, socType, socPath string, socPort int, protType string, shutCh chan struct{}) error {
	routes := map[string]util.HandlerFunc{
		authPath: func(w http.ResponseWriter, r *http.Request) error {
			return authPage(w, r, sys)
		},
		loginPath: func(w http.ResponseWriter, r *http.Request) error {
			return loginPage(w, r, sys)
		},
		selPath: func(w http.ResponseWriter, r *http.Request) error {
			return selectPage(w, r, sys)
		},
		consPath: func(w http.ResponseWriter, r *http.Request) error {
			return consentPage(w, r, sys)
		},
		tokPath: func(w http.ResponseWriter, r *http.Request) error {
			return tokenApi(w, r, sys)
		},
		accInfPath: func(w http.ResponseWriter, r *http.Request) error {
			return accountInfoApi(w, r, sys)
		},
	}
	if routes["/"] == nil {
		routes["/"] = func(w http.ResponseWriter, r *http.Request) error {
			return newIdpError(errInvReq, "invalid endpoint", http.StatusNotFound, nil)
		}
	}
	if sys.uiPath != "" {
		// ファイル配信も自前でやる。
		fileHndl := http.StripPrefix(sys.uiUri, http.FileServer(http.Dir(sys.uiPath)))
		for _, uri := range []string{sys.uiUri, sys.uiUri + "/"} {
			routes[uri] = func(w http.ResponseWriter, r *http.Request) error {
				fileHndl.ServeHTTP(w, r)
				return nil
			}
		}
	}
	return util.TerminableServe(socType, socPath, socPort, protType, routes, shutCh, panicErrorWrapper)
}
