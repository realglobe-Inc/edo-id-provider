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
	"testing"
)

func TestPasswordOnly(t *testing.T) {
	passwd := "ltFq9kclPgMK4ilaOF7fNlx2TE9OYFiyrX4x9gwCc9n"
	pass := newPasswordOnly(passwd)
	if pass.password() != passwd {
		t.Error(pass.password())
		t.Fatal(passwd)
	} else if len(pass.params()) > 0 {
		t.Fatal(pass.params())
	}
}
