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

package key

import (
	"encoding/json"
	cryptoutil "github.com/realglobe-Inc/edo-lib/crypto"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/go-lib/erro"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ファイルによる自身の鍵の格納庫。
type fileDb string

const (
	ext_Json = ".json"
	ext_Pem  = ".pem"
	ext_Key  = ".key"
	ext_Pub  = ".pub"
)

func NewFileDb(path string) Db {
	return fileDb(path)
}

func (this fileDb) Get() ([]jwk.Key, error) {
	dir, err := os.Open(string(this))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	keys := []jwk.Key{}
	for _, file := range files {
		if file.Size() == 0 {
			// ダミーファイル。
			continue
		}
		var newKeys []jwk.Key
		switch filepath.Ext(file.Name()) {
		case ext_Json:
			newKeys, err = this.readJson(filepath.Join(string(this), file.Name()))
		case ext_Pem, ext_Key, ext_Pub:
			newKeys, err = this.readPem(filepath.Join(string(this), file.Name()))
		}
		if err != nil {
			return nil, erro.Wrap(err)
		}
		keys = append(keys, newKeys...)
	}

	return keys, nil
}

func (this fileDb) readJson(path string) ([]jwk.Key, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	keys := []jwk.Key{}
	if data[0] == '[' {
		var ma []map[string]interface{}
		if err := json.Unmarshal(data, &ma); err != nil {
			return nil, erro.Wrap(err)
		}
		for _, m := range ma {
			key, err := jwk.FromMap(m)
			if err != nil {
				return nil, erro.Wrap(err)
			}
			keys = append(keys, key)
		}
	} else {
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, erro.Wrap(err)
		}
		key, err := jwk.FromMap(m)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (this fileDb) readPem(path string) ([]jwk.Key, error) {
	raws, err := cryptoutil.ReadPemAll(path)
	if err != nil {
		return nil, erro.Wrap(err)
	}

	keys := []jwk.Key{}
	for _, raw := range raws {
		key := jwk.New(raw, nil)
		if key == nil {
			continue
		}
		keys = append(keys, key)
	}
	return keys, nil
}
