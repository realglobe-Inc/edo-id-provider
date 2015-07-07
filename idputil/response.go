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

package idputil

import (
	"encoding/json"
	"net/http"

	"github.com/realglobe-Inc/go-lib/erro"
)

// 機密情報を JSON で返す。
func RespondJson(w http.ResponseWriter, params map[string]interface{}) error {
	buff, err := json.Marshal(params)
	if err != nil {
		return erro.Wrap(err)
	}

	w.Header().Add(tagContent_type, contTypeJson)
	w.Header().Add(tagCache_control, tagNo_store)
	w.Header().Add(tagPragma, tagNo_cache)
	if _, err := w.Write(buff); err != nil {
		log.Err(erro.Wrap(err))
	}
	return nil
}
