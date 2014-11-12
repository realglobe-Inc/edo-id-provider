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

	hndl := util.InitConsoleLog("github.com/realglobe-Inc")

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	hndl.SetLevel(param.consLv)
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

	var taExp driver.TaExplorer
	switch param.taExpType {
	case "file":
		taExp = driver.NewFileTaExplorer(param.taExpPath, 0)
		log.Info("Use file TA explorer " + param.taExpPath + ".")
	case "web":
		taExp = driver.NewWebTaExplorer(param.taExpAddr)
		log.Info("Use web TA explorer " + param.taExpAddr + ".")
	case "mongo":
		taExp, err = driver.NewMongoTaExplorer(param.taExpUrl, param.taExpDb, param.taExpColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb TA explorer " + param.taExpUrl + ".")
	default:
		return erro.New("invalid TA explorer type " + param.taExpType + ".")
	}

	var taKeyReg driver.TaKeyProvider
	switch param.taKeyRegType {
	case "file":
		taKeyReg = driver.NewFileTaKeyProvider(param.taKeyRegPath, 0)
		log.Info("Use file TA key provider " + param.taKeyRegPath + ".")
	case "web":
		taKeyReg = driver.NewWebTaKeyProvider(param.taKeyRegAddr)
		log.Info("Use web TA key provider " + param.taKeyRegAddr + ".")
	case "mongo":
		taKeyReg, err = driver.NewMongoTaKeyProvider(param.taKeyRegUrl, param.taKeyRegDb, param.taKeyRegColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb TA key provider " + param.taKeyRegUrl + ".")
	default:
		return erro.New("invalid TA key provider type " + param.taKeyRegType + ".")
	}

	var usrNameIdx driver.UserNameIndex
	switch param.usrNameIdxType {
	case "file":
		usrNameIdx = driver.NewFileUserNameIndex(param.usrNameIdxPath, 0)
		log.Info("Use file user name index " + param.usrNameIdxPath + ".")
	case "mongo":
		usrNameIdx, err = driver.NewMongoUserNameIndex(param.usrNameIdxUrl, param.usrNameIdxDb, param.usrNameIdxColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb user name index " + param.usrNameIdxUrl + ".")
	default:
		return erro.New("invalid user name index type " + param.usrNameIdxType + ".")
	}

	var usrAttrReg driver.UserAttributeRegistry
	switch param.usrAttrRegType {
	case "file":
		usrAttrReg = driver.NewFileUserAttributeRegistry(param.usrAttrRegPath, 0)
		log.Info("Use file user attribute registry " + param.usrAttrRegPath + ".")
	case "mongo":
		usrAttrReg, err = driver.NewMongoUserAttributeRegistry(param.usrAttrRegUrl, param.usrAttrRegDb, param.usrAttrRegColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb user attribute registry " + param.usrAttrRegUrl + ".")
	default:
		return erro.New("invalid user attribute registry type " + param.usrAttrRegType + ".")
	}

	jsonKeyGen := func(before string) string {
		return before + ".json"
	}

	var sessCont driver.TimeLimitedKeyValueStore
	switch param.sessContType {
	case "memory":
		sessCont = driver.NewMemoryTimeLimitedKeyValueStore(0)
		log.Info("Use memory session container.")
	case "file":
		sessCont = driver.NewFileTimeLimitedKeyValueStore(param.sessContPath, jsonKeyGen, json.Marshal, sessionUnmarshal, 0)
		log.Info("Use file session container " + param.sessContPath + ".")
	case "mongo":
		cont, err := driver.NewMongoTimeLimitedKeyValueStore(param.sessContUrl, param.sessContDb, param.sessContColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		cont.SetMongoTake(sessionMongoTake)
		sessCont = cont
		log.Info("Use mongodb session container " + param.sessContUrl + ".")
	default:
		return erro.New("invalid session container type " + param.sessContType + ".")
	}

	var codeCont driver.TimeLimitedKeyValueStore
	switch param.codeContType {
	case "memory":
		codeCont = driver.NewMemoryTimeLimitedKeyValueStore(0)
		log.Info("Use memory code container.")
	case "file":
		codeCont = driver.NewFileTimeLimitedKeyValueStore(param.codeContPath, jsonKeyGen, json.Marshal, codeUnmarshal, 0)
		log.Info("Use file code container " + param.codeContPath + ".")
	case "mongo":
		cont, err := driver.NewMongoTimeLimitedKeyValueStore(param.codeContUrl, param.codeContDb, param.codeContColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		cont.SetMongoTake(codeMongoTake)
		codeCont = cont
		log.Info("Use mongodb code container " + param.codeContUrl + ".")
	default:
		return erro.New("invalid code container type " + param.codeContType + ".")
	}

	var accTokenCont driver.TimeLimitedKeyValueStore
	switch param.accTokenContType {
	case "memory":
		accTokenCont = driver.NewMemoryTimeLimitedKeyValueStore(0)
		log.Info("Use memory access token container.")
	case "file":
		accTokenCont = driver.NewFileTimeLimitedKeyValueStore(param.accTokenContPath, jsonKeyGen, json.Marshal, accessTokenUnmarshal, 0)
		log.Info("Use file access token container " + param.accTokenContPath + ".")
	case "mongo":
		cont, err := driver.NewMongoTimeLimitedKeyValueStore(param.accTokenContUrl, param.accTokenContDb, param.accTokenContColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		cont.SetMongoTake(accessTokenMongoTake)
		accTokenCont = cont
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
