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
	"github.com/realglobe-Inc/edo-lib/strset"
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2/bson"
)

type Consent map[string]bool

func NewConsent(names ...string) Consent {
	cons := map[string]bool{}
	for _, name := range names {
		cons[name] = true
	}
	return cons
}

func (this Consent) Allow(name string) bool {
	return this[name]
}

func (this Consent) SetAllow(name string) {
	this[name] = true
}

func (this Consent) SetDeny(name string) {
	delete(this, name)
}

//  [
//      <許可項目>,
//      ...
//  ]
func (this Consent) GetBSON() (interface{}, error) {
	if this == nil {
		return nil, nil
	}

	return strset.Set(this), nil
}

func (this *Consent) SetBSON(raw bson.Raw) error {
	var buff strset.Set
	if err := raw.Unmarshal(&buff); err != nil {
		return erro.Wrap(err)
	}
	*this = Consent(buff)
	return nil
}
