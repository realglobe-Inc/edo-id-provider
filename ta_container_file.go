// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo-lib/driver"
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
