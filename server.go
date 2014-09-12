package main

import (
	"errors"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net"
	"net/http"
	"net/http/fcgi"
	"runtime"
)

var invalidProtocol = errors.New("invalid protocol.")

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

// / でリクエストを受け取って実行する。
func server(sys *system, lis net.Listener, routProtType string) error {
	var serv func(net.Listener, http.Handler) error
	switch routProtType {
	case "http":
		serv = http.Serve
	case "fcgi":
		serv = fcgi.Serve
	default:
		return erro.Wrap(invalidProtocol)
	}

	log.Debug("Server starts.")
	defer log.Debug("Server exits.")

	mux := http.NewServeMux()

	mux.HandleFunc(routPagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return routPage(sys, w, r)
	}))
	mux.HandleFunc(loginPagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return loginPage(sys, w, r)
	}))
	mux.HandleFunc(logoutPagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return logoutPage(sys, w, r)
	}))
	mux.HandleFunc(beginSessPagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return beginSessionPage(sys, w, r)
	}))
	mux.HandleFunc(delCookiePagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return delCookiePage(sys, w, r)
	}))
	mux.HandleFunc(setCookiePagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return setCookiePage(sys, w, r)
	}))

	mux.HandleFunc(accTokenPagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return accessTokenPage(sys, w, r)
	}))

	mux.HandleFunc(queryPagePath, panicErrorWrapper(func(w http.ResponseWriter, r *http.Request) error {
		return queryPage(sys, w, r)
	}))

	if err := serv(lis, mux); err != nil {
		return erro.Wrap(err)
	}

	return nil
}

// パニックとエラーの処理をまとめる。
func panicErrorWrapper(f func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// panic時にプロセス終了しないようにrecoverする
		defer func() {
			if rcv := recover(); rcv != nil {
				buff := make([]byte, 8192)
				stackLen := runtime.Stack(buff, false)
				stack := string(buff[:stackLen])
				err := erro.Wrap(util.NewPanicWrapper(rcv, stack))

				log.Err(erro.Unwrap(err))
				log.Debug(err)

				w.Header().Set("Content-Type", util.ContentTypeJson)
				http.Error(w, string(util.ErrorToResponseJson(err)), http.StatusInternalServerError)
				return
			}
		}()

		//////////////////////////////
		util.LogRequest(r, true)
		//////////////////////////////

		if err := f(w, r); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)

			var status int
			switch e := erro.Unwrap(err).(type) {
			case *util.HttpStatusError:
				status = e.Status()
			default:
				status = http.StatusInternalServerError
			}
			w.Header().Set("Content-Type", util.ContentTypeJson)
			http.Error(w, string(util.ErrorToResponseJson(err)), status)
			return
		}
	}
}
