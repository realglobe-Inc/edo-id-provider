package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/http"
)

func responseError(w http.ResponseWriter, statCod, errCod int, errDesc string) error {
	m := map[string]string{formErr: errCods[errCod]}
	if errDesc != "" {
		m[formErrDesc] = errDesc
	}
	buff, err := json.Marshal(m)
	if err != nil {
		return erro.Wrap(err)
	}

	w.WriteHeader(statCod)
	if _, err := w.Write(buff); err != nil {
		err = erro.Wrap(err)
		log.Err(erro.Unwrap(err))
		log.Debug(err)
	}
	return nil
}
