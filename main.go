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

	hndl := util.InitLog("github.com/realglobe-Inc")

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	hndl.SetLevel(param.consLv)
	if err := util.SetupLog("github.com/realglobe-Inc", param.logType, param.logLv, param.idpLogPath, param.fluAddr, param.idpFluTag); err != nil {
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

	var servExp driver.ServiceExplorer
	switch param.servExpType {
	case "file":
		servExp = driver.NewFileServiceExplorer(param.servExpPath, 0)
		log.Info("Use file service explorer " + param.servExpPath + ".")
	case "web":
		servExp = driver.NewWebServiceExplorer(param.servExpAddr)
		log.Info("Use web service explorer " + param.servExpAddr + ".")
	case "mongo":
		servExp, err = driver.NewMongoServiceExplorer(param.servExpUrl, param.servExpDb, param.servExpColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb service explorer " + param.servExpUrl + ".")
	default:
		return erro.New("invalid service explorer type " + param.servExpType + ".")
	}

	var servKeyReg driver.ServiceKeyRegistry
	switch param.servKeyRegType {
	case "file":
		servKeyReg = driver.NewFileServiceKeyRegistry(param.servKeyRegPath, 0)
		log.Info("Use file service key registry " + param.servKeyRegPath + ".")
	case "web":
		servKeyReg = driver.NewWebServiceKeyRegistry(param.servKeyRegAddr)
		log.Info("Use web service key registry " + param.servKeyRegAddr + ".")
	case "mongo":
		servKeyReg, err = driver.NewMongoServiceKeyRegistry(param.servKeyRegUrl, param.servKeyRegDb, param.servKeyRegColl, 0)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb service key registry " + param.servKeyRegUrl + ".")
	default:
		return erro.New("invalid service key registry type " + param.servKeyRegType + ".")
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
		servExp,
		servKeyReg,
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
	return serve(sys, param.idpSocType, param.idpSocPath, param.idpSocPort, param.idpProtType)
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
