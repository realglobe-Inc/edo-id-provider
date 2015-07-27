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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDb(t *testing.T) {
	path, err := ioutil.TempDir("", "edo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	if buff, err := json.Marshal(test_key.ToMap()); err != nil {
		t.Fatal(err)
	} else if err := ioutil.WriteFile(filepath.Join(path, "key.json"), buff, 0644); err != nil {
		t.Fatal(err)
	} else if buff, err := json.Marshal([]map[string]interface{}{
		test_sigKey.ToMap(),
		test_encKey.ToMap(),
		test_veriKey.ToMap(),
	}); err != nil {
		t.Fatal(err)
	} else if err := ioutil.WriteFile(filepath.Join(path, "keys.json"), buff, 0644); err != nil {
		t.Fatal(err)
	}

	testDb(t, NewFileDb(path))
}

func TestFileDbReadPem(t *testing.T) {
	path, err := ioutil.TempDir("", "edo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	if err := ioutil.WriteFile(filepath.Join(path, "key.pem"), []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEOCbnPn2SPA92u2G09XmrB9rTeqWv
SFeYEjDv3p7hDnDS+vrPmEQ3twGw7vn38JoIIhYdowJX4+deWcezFDtI1A==
-----END PUBLIC KEY-----`), 0644); err != nil {
		t.Fatal(err)
	}

	db := NewFileDb(path)
	if keys, err := db.Get(); err != nil {
		t.Fatal(err)
	} else if len(keys) != 1 {
		t.Fatal(keys)
	} else if keys[0].Type() != "EC" {
		t.Fatal(keys[0])
	}
}
