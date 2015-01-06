package main

import (
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"github.com/realglobe-Inc/go-lib-rg/rglog"
	"net/http"
	"os"
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

	var sessCont sessionContainer
	switch param.sessContType {
	case "memory":
		sessCont = newMemorySessionContainer(param.sessIdLen, param.sessExpiDur, param.caStaleDur, param.caExpiDur)
		log.Info("Use memory session container.")
	case "file":
		sessCont = newFileSessionContainer(param.sessIdLen, param.sessExpiDur, param.sessContPath, param.sessExpiContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file session container " + param.sessContPath + "," + param.sessExpiContPath)
	case "redis":
		sessCont = newRedisSessionContainer(param.sessIdLen, param.sessExpiDur, param.sessContUrl, param.sessContPrefix, param.caStaleDur, param.caExpiDur)
		log.Info("Use redis session container " + param.sessContUrl)
	default:
		return erro.New("invalid session container type " + param.sessContType)
	}

	var codCont codeContainer
	switch param.codContType {
	case "memory":
		codCont = newMemoryCodeContainer(param.codIdLen, param.codExpiDur, param.caStaleDur, param.caExpiDur)
		log.Info("Use memory code container.")
	case "file":
		codCont = newFileCodeContainer(param.codIdLen, param.codExpiDur, param.codContPath, param.codExpiContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file code container " + param.codContPath + "," + param.codExpiContPath)
	case "redis":
		codCont = newRedisCodeContainer(param.codIdLen, param.codExpiDur, param.codContUrl, param.codContPrefix, param.caStaleDur, param.caExpiDur)
		log.Info("Use mongodb code container " + param.codContUrl)
	default:
		return erro.New("invalid code container type " + param.codContType)
	}

	var tokCont tokenContainer
	switch param.tokContType {
	case "memory":
		tokCont = newMemoryTokenContainer(param.tokIdLen, param.tokExpiDur, param.maxTokExpiDur, param.caStaleDur, param.caExpiDur)
		log.Info("Use memory token container.")
	case "file":
		tokCont = newFileTokenContainer(param.tokIdLen, param.tokExpiDur, param.maxTokExpiDur, param.tokContPath, param.tokExpiContPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file token container " + param.tokContPath + "," + param.tokExpiContPath)
	case "redis":
		tokCont = newRedisTokenContainer(param.tokIdLen, param.tokExpiDur, param.maxTokExpiDur, param.tokContUrl, param.tokContPrefix, param.caStaleDur, param.caExpiDur)
		log.Info("Use mongodb token container " + param.tokContUrl)
	default:
		return erro.New("invalid token container type " + param.tokContType)
	}

	sys := newSystem(
		param.selfId,
		param.secCook,
		param.codIdLen/2,
		param.codIdLen/2,
		param.uiUri,
		param.uiPath,
		taCont,
		accCont,
		sessCont,
		codCont,
		tokCont,
		param.tokExpiDur,
	)
	return serve(sys, param.socType, param.socPath, param.socPort, param.protType)
}

// 振り分ける。
const (
	authPath   = "/login"
	tokPath    = "/token"
	accInfPath = "/account"
)

func serve(sys *system, socType, socPath string, socPort int, protType string) error {
	routes := map[string]util.HandlerFunc{
		authPath: func(w http.ResponseWriter, r *http.Request) error {
			return authPage(sys, w, r)
		},
		tokPath: func(w http.ResponseWriter, r *http.Request) error {
			return tokenApi(sys, w, r)
		},
		accInfPath: func(w http.ResponseWriter, r *http.Request) error {
			return accountInfoApi(sys, w, r)
		},
	}
	fileHndl := http.StripPrefix(sys.uiUri, http.FileServer(http.Dir(sys.uiPath)))
	for _, uri := range []string{sys.uiUri, sys.uiUri + "/"} {
		routes[uri] = func(w http.ResponseWriter, r *http.Request) error {
			fileHndl.ServeHTTP(w, r)
			return nil
		}
	}
	return util.Serve(socType, socPath, socPort, protType, routes)
}
