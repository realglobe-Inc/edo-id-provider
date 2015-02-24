package main

import (
	"github.com/realglobe-Inc/go-lib/erro"
	"net/http"
)

func redirectError(w http.ResponseWriter, r *http.Request, sys *system, sess *session, authReq *authRequest, err error) error {
	if sess != nil && sess.id() != "" {
		// 認証経過を廃棄。
		sess.abort()
		if err := sys.sessCont.put(sess); err != nil {
			err = erro.Wrap(err)
			log.Err(erro.Unwrap(err))
			log.Debug(err)
		} else {
			log.Debug("Session " + mosaic(sess.id()) + " was aborted")
		}
	}

	q := authReq.redirectUri().Query()
	switch e := erro.Unwrap(err).(type) {
	case *idpError:
		log.Err(e.errorDescription())
		log.Debug(e)
		q.Set(formErr, e.errorCode())
		q.Set(formErrDesc, e.errorDescription())
	default:
		log.Err(e)
		log.Debug(err)
		q.Set(formErr, errServErr)
		q.Set(formErrDesc, e.Error())
	}
	if authReq.state() != "" {
		q.Set(formStat, authReq.state())
	}

	authReq.redirectUri().RawQuery = q.Encode()
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	http.Redirect(w, r, authReq.redirectUri().String(), http.StatusFound)
	return nil
}
