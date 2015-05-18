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

package request

import (
	"reflect"
	"testing"
)

func TestFormValueSet(t *testing.T) {
	if s, s2 := map[string]bool{"openid": true, "email": true}, FormValueSet("openid email"); !reflect.DeepEqual(s2, s) {
		t.Error(s2)
		t.Fatal(s)
	}
}

func TestValueSetForm(t *testing.T) {
	if f := ValueSetForm(map[string]bool{"openid": true, "email": true}); f != "openid email" && f != "email openid" {
		t.Error(f)
		t.Fatal(`"openid email" or "email openid"`)
	}
}

func TestFormValues(t *testing.T) {
	if s, s2 := []string{"openid", "email"}, FormValues("openid email"); !reflect.DeepEqual(s2, s) {
		t.Error(s2)
		t.Fatal(s)
	}
}

func TestValuesForm(t *testing.T) {
	if f := ValuesForm([]string{"openid", "email"}); f != "openid email" {
		t.Error(f)
		t.Fatal(`"openid email"`)
	}
}
