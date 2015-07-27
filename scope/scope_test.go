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

package scope

import (
	"reflect"
	"testing"

	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
)

func TestRemoveUnknown(t *testing.T) {
	base := []string{"openid", "email"}
	scop1 := strsetutil.New(base...)
	scop2 := strsetutil.New(base...)

	scop2["unknown"] = true
	scop2 = RemoveUnknown(scop2)

	if !reflect.DeepEqual(scop2, scop1) {
		t.Error(scop2)
		t.Fatal(scop1)
	}
}

func TestAttributes(t *testing.T) {
	attrs1 := strsetutil.New("email", "email_verified")

	if attrs2 := Attributes("email"); !reflect.DeepEqual(attrs2, attrs1) {
		t.Error(attrs2)
		t.Fatal(attrs1)
	}
}

func TestFromAttribute(t *testing.T) {
	if scop := FromAttribute("email_verified"); scop != "email" {
		t.Error(scop)
		t.Fatal("email")
	}
}

func TestIsEssential(t *testing.T) {
	if !IsEssential("openid") {
		t.Fatal("openid is not essential")
	} else if IsEssential("email") {
		t.Fatal("email is essential")
	}
}
