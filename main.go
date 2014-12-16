package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
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
	var taExp TaExplorer
	switch param.taExpType {
	case "file":
		taExp = NewFileTaExplorer(param.taExpPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file TA explorer " + param.taExpPath + ".")
	default:
		return erro.New("invalid TA explorer type " + param.taExpType + ".")
	}

	var taKeyReg TaKeyProvider
	switch param.taKeyRegType {
	case "file":
		taKeyReg = NewFileTaKeyProvider(param.taKeyRegPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file TA key provider " + param.taKeyRegPath + ".")
	default:
		return erro.New("invalid TA key provider type " + param.taKeyRegType + ".")
	}

	var usrNameIdx UserNameIndex
	switch param.usrNameIdxType {
	case "file":
		usrNameIdx = NewFileUserNameIndex(param.usrNameIdxPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file user name index " + param.usrNameIdxPath + ".")
	default:
		return erro.New("invalid user name index type " + param.usrNameIdxType + ".")
	}

	var usrAttrReg UserAttributeRegistry
	switch param.usrAttrRegType {
	case "file":
		usrAttrReg = NewFileUserAttributeRegistry(param.usrAttrRegPath, param.caStaleDur, param.caExpiDur)
		log.Info("Use file user attribute registry " + param.usrAttrRegPath + ".")
	default:
		return erro.New("invalid user attribute registry type " + param.usrAttrRegType + ".")
	}

	var sessCont driver.TimeLimitedKeyValueStore
	switch param.sessContType {
	case "memory":
		sessCont = driver.NewMemoryTimeLimitedKeyValueStore(param.caStaleDur, param.caExpiDur)
		log.Info("Use memory session container.")
	case "file":
		sessCont = driver.NewFileTimeLimitedKeyValueStore(param.sessContPath, param.sessContPath+".expi", keyToJsonPath, jsonPathToKey, json.Marshal, sessionUnmarshal, param.caStaleDur, param.caExpiDur)
		log.Info("Use file session container " + param.sessContPath + ".")
	default:
		return erro.New("invalid session container type " + param.sessContType + ".")
	}

	var codeCont driver.TimeLimitedKeyValueStore
	switch param.codeContType {
	case "memory":
		codeCont = driver.NewMemoryTimeLimitedKeyValueStore(param.caStaleDur, param.caExpiDur)
		log.Info("Use memory code container.")
	case "file":
		codeCont = driver.NewFileTimeLimitedKeyValueStore(param.codeContPath, param.codeContPath+".expi", keyToJsonPath, jsonPathToKey, json.Marshal, codeUnmarshal, param.caStaleDur, param.caExpiDur)
		log.Info("Use file code container " + param.codeContPath + ".")
	default:
		return erro.New("invalid code container type " + param.codeContType + ".")
	}

	var accTokenCont driver.TimeLimitedKeyValueStore
	switch param.accTokenContType {
	case "memory":
		accTokenCont = driver.NewMemoryTimeLimitedKeyValueStore(param.caStaleDur, param.caExpiDur)
		log.Info("Use memory access token container.")
	case "file":
		accTokenCont = driver.NewFileTimeLimitedKeyValueStore(param.accTokenContPath, param.accTokenContPath+".expi", keyToJsonPath, jsonPathToKey, json.Marshal, accessTokenUnmarshal, param.caStaleDur, param.caExpiDur)
		log.Info("Use file access token container " + param.accTokenContPath + ".")
	default:
		return erro.New("invalid access token container type " + param.accTokenContType + ".")
	}

	sys := &system{
		taExp,
		taKeyReg,
		usrNameIdx,
		usrAttrReg,
		sessCont,
		codeCont,
		accTokenCont,
		param.maxSessExpiDur,
		param.codeExpiDur,
		param.accTokenExpiDur,
		param.maxAccTokenExpiDur,
	}
	return serve(sys, param.socType, param.socPath, param.socPort, param.protType)
}

// 振り分ける。
const (
	routPagePath      = "/"
	loginPagePath     = "/login"
	logoutPagePath    = "/logout"
	beginSessPagePath = "/begin_session"
	setCookiePagePath = "/set_cookie"
	delCookiePagePath = "/delete_cookie"

	accTokenPagePath = "/access_token"

	queryPagePath = "/query"
)

func serve(sys *system, socType, socPath string, socPort int, protType string) error {
	routes := map[string]util.HandlerFunc{
		routPagePath: func(w http.ResponseWriter, r *http.Request) error {
			return routPage(sys, w, r)
		},
		loginPagePath: func(w http.ResponseWriter, r *http.Request) error {
			return loginPage(sys, w, r)
		},
		logoutPagePath: func(w http.ResponseWriter, r *http.Request) error {
			return logoutPage(sys, w, r)
		},
		beginSessPagePath: func(w http.ResponseWriter, r *http.Request) error {
			return beginSessionPage(sys, w, r)
		},
		delCookiePagePath: func(w http.ResponseWriter, r *http.Request) error {
			return deleteCookiePage(sys, w, r)
		},
		setCookiePagePath: func(w http.ResponseWriter, r *http.Request) error {
			return setCookiePage(sys, w, r)
		},
		accTokenPagePath: func(w http.ResponseWriter, r *http.Request) error {
			return accessTokenPage(sys, w, r)
		},
		queryPagePath: func(w http.ResponseWriter, r *http.Request) error {
			return queryPage(sys, w, r)
		},
	}
	return util.Serve(socType, socPath, socPort, protType, routes)
}
