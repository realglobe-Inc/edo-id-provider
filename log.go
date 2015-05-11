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
	"github.com/realglobe-Inc/go-lib/rglog"
)

var log = rglog.Logger("github.com/realglobe-Inc/edo-id-provider")

const mosaicThres = 10

// ログにそのまま書くのが憚られるので隠す。
func mosaic(str string) string {
	if len(str) <= mosaicThres {
		return str
	} else {
		return str[:mosaicThres] + "..."
	}
}
