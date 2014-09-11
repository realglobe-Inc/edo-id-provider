package main

import (
	"fmt"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"github.com/realglobe-Inc/go-lib-rg/rglog"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	hndl := util.InitLog("github.com/realglobe-Inc")

	param, err := parseParameters(os.Args...)
	if err != nil {
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	hndl.SetLevel(param.consLv)

	switch param.logType {
	case "":
	case "file":
		if err := util.InitFileLog("github.com/realglobe-Inc", param.logLv, param.idpLogPath); err != nil {
			log.Err(erro.Unwrap(err))
			log.Debug(err)
			exitCode = 1
			return
		}
	case "fluentd":
		if err := util.InitFluentdLog("github.com/realglobe-Inc", param.logLv, param.fluAddr, param.idpFluTag); err != nil {
			log.Err(erro.Unwrap(err))
			log.Debug(err)
			exitCode = 1
			return
		}
	default:
		log.Err("Invalid log type: " + param.logType + ".")
		log.Debug(err)
		exitCode = 1
		return
	}

	shutCh := make(chan struct{}, 1)

	// SIGINT、SIGKILL、SIGTERM を受け取ったら終了。
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Info("Signal ", sig, " was caught.")
		shutCh <- struct{}{}
	}()

	if err := mainCore(shutCh, param); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
		exitCode = 1
		return
	}

	log.Info("Shut down.")
}

func mainCore(shutCh chan struct{}, param *parameters) error {
	var err error

	var servExp driver.ServiceExplorer
	switch param.servExpType {
	case "file":
		servExp = driver.NewFileServiceExplorer(param.servExpPath)
		log.Info("Use file service explorer " + param.servExpPath + ".")
	case "web":
		servExp = driver.NewWebServiceExplorer(param.servExpAddr)
		log.Info("Use web service explorer " + param.servExpAddr + ".")
	case "mongo":
		servExp, err = driver.NewMongoServiceExplorer(param.servExpUrl, param.servExpDb, param.servExpColl)
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb service explorer " + param.servExpUrl + ".")
	default:
		return erro.New("invalid service explorer type " + param.servExpType + ".")
	}

	var usrNameIdx driver.UserNameIndex
	switch param.usrNameIdxType {
	case "file":
		usrNameIdx = driver.NewFileUserNameIndex(param.usrNameIdxPath)
		log.Info("Use file user name index " + param.usrNameIdxPath + ".")
	case "mongo":
		usrNameIdx, err = driver.NewMongoUserNameIndex(param.usrNameIdxUrl, param.usrNameIdxDb, param.usrNameIdxColl)
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
		usrAttrReg = driver.NewFileUserAttributeRegistry(param.usrAttrRegPath)
		log.Info("Use file user attribute registry " + param.usrAttrRegPath + ".")
	case "mongo":
		usrAttrReg, err = driver.NewMongoUserAttributeRegistry(param.usrAttrRegUrl, param.usrAttrRegDb, param.usrAttrRegColl)
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
		sessCont = driver.NewMemoryTimeLimitedKeyValueStore()
		log.Info("Use memory session container.")
	case "file":
		sessCont = driver.NewFileTimeLimitedKeyValueStore(param.sessContPath)
		log.Info("Use file session container " + param.sessContPath + ".")
	case "mongo":
		sessCont, err = driver.NewMongoTimeLimitedKeyValueStore(param.sessContUrl, param.sessContDb, param.sessContColl, "session_id", "session")
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb session container " + param.sessContUrl + ".")
	default:
		return erro.New("invalid session container type " + param.sessContType + ".")
	}

	var codeCont driver.TimeLimitedKeyValueStore
	switch param.codeContType {
	case "memory":
		codeCont = driver.NewMemoryTimeLimitedKeyValueStore()
		log.Info("Use memory code container.")
	case "file":
		codeCont = driver.NewFileTimeLimitedKeyValueStore(param.codeContPath)
		log.Info("Use file code container " + param.codeContPath + ".")
	case "mongo":
		codeCont, err = driver.NewMongoTimeLimitedKeyValueStore(param.codeContUrl, param.codeContDb, param.codeContColl, "code_id", "code")
		if err != nil {
			return erro.Wrap(err)
		}
		log.Info("Use mongodb code container " + param.codeContUrl + ".")
	default:
		return erro.New("invalid code container type " + param.codeContType + ".")
	}

	var sleepTime time.Duration = 0
	resetInterval := time.Minute
	for {
		if brk, err := func() (brk bool, err error) {
			var lis net.Listener
			defer func() {
				if lis != nil {
					lis.Close()
				}
			}()

			switch param.idpSocType {
			case "unix":
				lis, err = net.Listen("unix", param.idpSocPath)
				if err != nil {
					return false, erro.Wrap(err)
				}
				if err := os.Chmod(param.idpSocPath, 0777); err != nil {
					return false, erro.Wrap(err)
				}
				log.Info("Wait on UNIX socket " + param.idpSocPath + ".")
			case "tcp":
				lis, err = net.Listen("tcp", fmt.Sprint(":", param.idpSocPort))
				if err != nil {
					return false, erro.Wrap(err)
				}
				log.Info("Wait on TCP socket ", param.idpSocPort, ".")
			default:
				return true, erro.New("invalid socket type " + param.idpSocType + ".")
			}

			stopCh := make(chan struct{}, 1)
			subShutCh := make(chan bool, 1)
			go func() {
				select {
				case <-shutCh:
					subShutCh <- true
					lis.Close()
				case <-stopCh:
					subShutCh <- false
				}
			}()
			defer func() { stopCh <- struct{}{} }()

			sys := &system{
				servExp,
				usrNameIdx,
				usrAttrReg,
				sessCont,
				codeCont,
				param.cookieMaxAge,
				param.maxSessExpiDur,
				param.codeExpiDur,
			}

			start := time.Now()
			if err := server(sys, lis, param.idpProtType); err != nil {
				err := erro.Wrap(err)

				// 正常な終了処理としてソケットが閉じられたかもしれないので調べる。
				select {
				case <-subShutCh:
					return true, nil
				default:
				}

				stopCh <- struct{}{}
				brk = <-subShutCh

				if brk || erro.Unwrap(err) == invalidProtocol {
					// どうしようもない。
					return true, err
				}

				end := time.Now()
				if end.Sub(start) > resetInterval {
					sleepTime = 0
				}

				return false, err
			}

			stopCh <- struct{}{}
			return <-subShutCh, nil
		}(); brk {
			return erro.Wrap(err)
		} else {
			if err != nil {
				log.Err(erro.Unwrap(err))
				log.Debug(err)
			}

			sleepTime = nextSleepTime(sleepTime, resetInterval)
			log.Info("Retry after ", sleepTime)
			time.Sleep(sleepTime)
		}
	}

	return nil
}

func nextSleepTime(cur, max time.Duration) time.Duration {
	next := 2*cur + time.Duration(rand.Int63n(int64(time.Second)))
	if next >= max {
		next = time.Minute
	}
	return next
}
