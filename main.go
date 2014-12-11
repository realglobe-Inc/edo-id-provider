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
	var err error

	var taExp TaExplorer
	switch param.taExpType {
	case "file":
		taExp = NewFileTaExplorer(param.taExpPath, 0)
		log.Info("Use file TA explorer " + param.taExpPath + ".")
	case "web":
		taExp = NewWebTaExplorer(param.taExpAddr)
		log.Info("Use web TA explorer " + param.taExpAddr + ".")
	case "mongo":
		taExp, err = NewMongoTaExplorer(param.taExpUrl, param.taExpDb, param.taExpColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb TA explorer " + param.taExpUrl + ".")
	default:
		return erro.New("invalid TA explorer type " + param.taExpType + ".")
	}

	var taKeyReg TaKeyProvider
	switch param.taKeyRegType {
	case "file":
		taKeyReg = NewFileTaKeyProvider(param.taKeyRegPath, 0)
		log.Info("Use file TA key provider " + param.taKeyRegPath + ".")
	case "web":
		taKeyReg = NewWebTaKeyProvider(param.taKeyRegAddr)
		log.Info("Use web TA key provider " + param.taKeyRegAddr + ".")
	case "mongo":
		taKeyReg, err = NewMongoTaKeyProvider(param.taKeyRegUrl, param.taKeyRegDb, param.taKeyRegColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb TA key provider " + param.taKeyRegUrl + ".")
	default:
		return erro.New("invalid TA key provider type " + param.taKeyRegType + ".")
	}

	var usrNameIdx UserNameIndex
	switch param.usrNameIdxType {
	case "file":
		usrNameIdx = NewFileUserNameIndex(param.usrNameIdxPath, 0)
		log.Info("Use file user name index " + param.usrNameIdxPath + ".")
	case "mongo":
		usrNameIdx, err = NewMongoUserNameIndex(param.usrNameIdxUrl, param.usrNameIdxDb, param.usrNameIdxColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb user name index " + param.usrNameIdxUrl + ".")
	default:
		return erro.New("invalid user name index type " + param.usrNameIdxType + ".")
	}

	var usrAttrReg UserAttributeRegistry
	switch param.usrAttrRegType {
	case "file":
		usrAttrReg = NewFileUserAttributeRegistry(param.usrAttrRegPath, 0)
		log.Info("Use file user attribute registry " + param.usrAttrRegPath + ".")
	case "mongo":
		usrAttrReg, err = NewMongoUserAttributeRegistry(param.usrAttrRegUrl, param.usrAttrRegDb, param.usrAttrRegColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb user attribute registry " + param.usrAttrRegUrl + ".")
	default:
		return erro.New("invalid user attribute registry type " + param.usrAttrRegType + ".")
	}

	var sessCont driver.TimeLimitedKeyValueStore
	switch param.sessContType {
	case "memory":
		sessCont = driver.NewMemoryTimeLimitedKeyValueStore(0, 0)
		log.Info("Use memory session container.")
	case "file":
		sessCont = driver.NewFileTimeLimitedKeyValueStore(param.sessContPath, keyToJsonPath, jsonPathToKey, json.Marshal, sessionUnmarshal, 0, 0)
		log.Info("Use file session container " + param.sessContPath + ".")
	case "mongo":
		sessCont = driver.NewMongoTimeLimitedKeyValueStore(param.sessContUrl, param.sessContDb, param.sessContColl, nil, nil, sessionMongoTake, 0, 0)
		log.Info("Use mongodb session container " + param.sessContUrl + ".")
	default:
		return erro.New("invalid session container type " + param.sessContType + ".")
	}

	var codeCont driver.TimeLimitedKeyValueStore
	switch param.codeContType {
	case "memory":
		codeCont = driver.NewMemoryTimeLimitedKeyValueStore(0, 0)
		log.Info("Use memory code container.")
	case "file":
		codeCont = driver.NewFileTimeLimitedKeyValueStore(param.codeContPath, keyToJsonPath, jsonPathToKey, json.Marshal, codeUnmarshal, 0, 0)
		log.Info("Use file code container " + param.codeContPath + ".")
	case "mongo":
		codeCont = driver.NewMongoTimeLimitedKeyValueStore(param.codeContUrl, param.codeContDb, param.codeContColl, nil, nil, codeMongoTake, 0, 0)
		log.Info("Use mongodb code container " + param.codeContUrl + ".")
	default:
		return erro.New("invalid code container type " + param.codeContType + ".")
	}

	var accTokenCont driver.TimeLimitedKeyValueStore
	switch param.accTokenContType {
	case "memory":
		accTokenCont = driver.NewMemoryTimeLimitedKeyValueStore(0, 0)
		log.Info("Use memory access token container.")
	case "file":
		accTokenCont = driver.NewFileTimeLimitedKeyValueStore(param.accTokenContPath, keyToJsonPath, jsonPathToKey, json.Marshal, accessTokenUnmarshal, 0, 0)
		log.Info("Use file access token container " + param.accTokenContPath + ".")
	case "mongo":
		accTokenCont = driver.NewMongoTimeLimitedKeyValueStore(param.accTokenContUrl, param.accTokenContDb, param.accTokenContColl, nil, nil, accessTokenMongoTake, 0, 0)
		log.Info("Use mongodb access token container " + param.accTokenContUrl + ".")
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
