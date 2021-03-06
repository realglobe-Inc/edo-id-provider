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

package consent

import (
	"testing"
)

func TestConsent(t *testing.T) {
	cons := NewConsent()
	if cons.Allow(test_scop) {
		t.Fatal("allow not allowed")
	}

	cons.SetAllow(test_scop)
	if !cons.Allow(test_scop) {
		t.Fatal("deny allowed")
	}

	cons.SetDeny(test_scop)
	if cons.Allow(test_scop) {
		t.Fatal("allow denied")
	}
}
