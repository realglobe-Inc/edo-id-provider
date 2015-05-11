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
	idperr "github.com/realglobe-Inc/edo-idp-selector/error"
	"github.com/realglobe-Inc/go-lib/erro"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestResponse(t *testing.T) {
	m := map[string]interface{}{
		"access_token": "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC",
		"scope":        "openid email",
		"expires_in":   3600.0,
		"token_type":   "Bearer",
	}

	w := httptest.NewRecorder()
	if err := response(w, m); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusOK {
		t.Error(w.Code)
		t.Fatal(http.StatusOK)
	} else if w.HeaderMap.Get("Content-Type") != "application/json" {
		t.Error(w.HeaderMap.Get("Content-Type"))
		t.Fatal("application/json")
	} else if w.Body == nil {
		t.Fatal("no body")
	}

	data, _ := ioutil.ReadAll(w.Body)
	var m2 map[string]interface{}
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(m2, m) {
		t.Error(m2)
		t.Fatal(m2)
	}
}

func TestResponseIdpError(t *testing.T) {
	origErr := idperr.New(idperr.Invalid_request, "invalid request", http.StatusBadRequest, nil)

	w := httptest.NewRecorder()
	if err := responseError(w, erro.Wrap(origErr)); err != nil {
		t.Fatal(err)
	}

	if w.Code != origErr.Status() {
		t.Error(w.Code)
		t.Fatal(origErr.Status())
	} else if w.HeaderMap.Get("Content-Type") != "application/json" {
		t.Error(w.HeaderMap.Get("Content-Type"))
		t.Fatal("application/json")
	} else if w.Body == nil {
		t.Fatal("no body")
	}

	data, _ := ioutil.ReadAll(w.Body)
	var buff struct {
		Error             string
		Error_description string
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		t.Fatal(err)
	} else if buff.Error != origErr.ErrorCode() {
		t.Error(buff.Error)
		t.Fatal(origErr.ErrorCode())
	} else if buff.Error_description != origErr.ErrorDescription() {
		t.Error(buff.Error_description)
		t.Fatal(origErr.ErrorDescription())
	}
}

func TestResponseNormalError(t *testing.T) {
	origErr := erro.New("test error")

	w := httptest.NewRecorder()
	if err := responseError(w, origErr); err != nil {
		t.Fatal(err)
	}

	if w.Code != http.StatusInternalServerError {
		t.Error(w.Code)
		t.Fatal(http.StatusInternalServerError)
	} else if w.HeaderMap.Get("Content-Type") != "application/json" {
		t.Error(w.HeaderMap.Get("Content-Type"))
		t.Fatal("application/json")
	} else if w.Body == nil {
		t.Fatal("no body")
	}

	data, _ := ioutil.ReadAll(w.Body)
	var buff struct {
		Error             string
		Error_description string
	}
	if err := json.Unmarshal(data, &buff); err != nil {
		t.Fatal(err)
	} else if buff.Error != idperr.Server_error {
		t.Error(buff.Error)
		t.Fatal(idperr.Server_error)
	} else if buff.Error_description != "test error" {
		t.Error(buff.Error_description)
		t.Fatal("test error")
	}
}
