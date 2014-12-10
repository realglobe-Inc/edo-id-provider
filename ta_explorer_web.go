package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/url"
)

// {
//     "service": {
//         "uuid": "service-no-uuid"
//     }
// }
func webServiceUuidUnmarshal(data []byte) (interface{}, error) {
	var res struct {
		Service struct {
			Uuid string
		}
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return res.Service.Uuid, nil
}

type webTaExplorer struct {
	base driver.KeyValueStore
}

// スレッドセーフ。
func NewWebTaExplorer(prefix string) TaExplorer {
	return &webTaExplorer{driver.NewWebKeyValueStore(prefix, nil, webServiceUuidUnmarshal)}
}

func (reg *webTaExplorer) ServiceUuid(servUri string, caStmp *driver.Stamp) (servUuid string, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := reg.base.Get("?service_uri="+url.QueryEscape(servUri), caStmp)
	if err != nil {
		return "", nil, erro.Wrap(err)
	} else if value == nil || value == "" {
		return "", newCaStmp, nil
	}
	return value.(string), newCaStmp, nil
}
