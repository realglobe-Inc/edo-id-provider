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
