package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/url"
	"path/filepath"
	"time"
)

func unmarshalConsent(data []byte) (interface{}, error) {
	var cons consent
	if err := json.Unmarshal(data, &cons); err != nil {
		return nil, erro.Wrap(err)
	}
	return &cons, nil
}

// スレッドセーフ。
func newFileConsentContainer(path string, staleDur, expiDur time.Duration) consentContainer {
	return &consentContainerImpl{
		driver.NewFileListedKeyValueStore(path, keyToJsonPath, nil, json.Marshal, unmarshalConsent, staleDur, expiDur),
		func(accId, taId string) string {
			return filepath.Join(url.QueryEscape(accId), url.QueryEscape(taId))
		},
	}
}
