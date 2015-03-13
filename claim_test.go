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
	"fmt"
	"reflect"
	"testing"
)

func TestClaimRequest(t *testing.T) {
	req := claimRequest{
		"name":     {"ja": {true, nil, nil}},
		"nickname": {"": {false, nil, nil}},
	}

	buff, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var req2 claimRequest
	if err := json.Unmarshal(buff, &req2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(req2, req) {
		t.Error(fmt.Sprintf("%#v", req2))
		t.Error(fmt.Sprintf("%#v", req))
		t.Error(string(buff))
	}
}
