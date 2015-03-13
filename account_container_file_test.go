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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileAccountContainer(t *testing.T) {
	path, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	namePath, err := ioutil.TempDir("", testLabel)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(namePath)

	if buff, err := json.Marshal(testAcc); err != nil {
		t.Fatal(err)
	} else if err := ioutil.WriteFile(filepath.Join(path, keyToEscapedJsonPath(testAcc.id())), buff, filePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(path, keyToEscapedJsonPath(testAcc.id())), filepath.Join(namePath, keyToEscapedJsonPath(testAcc.name()))); err != nil {
		t.Fatal(err)
	}

	testAccountContainer(t, newFileAccountContainer(path, namePath, testStaleDur, testCaExpiDur))
}
