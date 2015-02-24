package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-toolkit/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"net/url"
	"strings"
	"time"
)

func unmarshalTa(data []byte) (interface{}, error) {
	var t ta
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, erro.Wrap(err)
	}
	return &t, nil
}

func keyToJsonPath(key string) string {
	return key + ".json"
}

func jsonPathToKey(path string) string {
	if !strings.HasSuffix(path, ".json") {
		return ""
	}
	return path[:len(path)-len(".json")]
}

func keyToEscapedJsonPath(key string) string {
	return keyToJsonPath(url.QueryEscape(key))
}

func escapedJsonPathToKey(path string) string {
	key, _ := url.QueryUnescape(jsonPathToKey(path))
	return key
}

// スレッドセーフ。
func newFileTaContainer(path string, staleDur, expiDur time.Duration) taContainer {
	return &taContainerImpl{driver.NewFileListedKeyValueStore(path,
		keyToEscapedJsonPath, escapedJsonPathToKey,
		json.Marshal, unmarshalTa,
		staleDur, expiDur)}
}
