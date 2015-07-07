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
	"testing"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	keydb "github.com/realglobe-Inc/edo-id-provider/database/key"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
)

type idTokenSystem struct {
	SetSubSystem
	keyDb    keydb.Db
	sigAlg   string
	sigKid   string
	selfId   string
	jtiExpIn time.Duration
}

func (this *idTokenSystem) KeyDb() keydb.Db               { return this.keyDb }
func (this *idTokenSystem) SignAlgorithm() string         { return this.sigAlg }
func (this *idTokenSystem) SignKeyId() string             { return this.sigKid }
func (this *idTokenSystem) SelfId() string                { return this.selfId }
func (this *idTokenSystem) JwtIdExpiresIn() time.Duration { return this.jtiExpIn }

func newIdTokenSystem(setSubSys SetSubSystem, key jwk.Key, sigAlg, selfId string) IdTokenSystem {
	return &idTokenSystem{
		setSubSys,
		keydb.NewMemoryDb([]jwk.Key{key}),
		sigAlg,
		"",
		selfId,
		time.Minute,
	}
}

func TestIdToken(t *testing.T) {
	key, err := jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})
	sys := newIdTokenSystem(newSetSubSystem(), key, "ES256", test_idpId)
	ta := tadb.New(test_taId, nil, nil, nil, false, test_sectId)
	acnt := account.New(test_acntId, test_acntName, nil, map[string]interface{}{"email": test_email})
	rawIdTok, err := IdToken(sys, ta, acnt, strsetutil.New("email"), map[string]interface{}{"aaaaa": "bbbbb"})
	if err != nil {
		t.Fatal(err)
	}
	idTok, err := jwt.Parse([]byte(rawIdTok))
	if err != nil {
		t.Fatal(err)
	} else if !idTok.IsSigned() {
		t.Fatal("not signed")
	} else if err := idTok.Verify([]jwk.Key{key}); err != nil {
		t.Fatal(err)
	} else if iss, _ := idTok.Claim("iss").(string); iss != test_idpId {
		t.Error(iss)
		t.Fatal(test_idpId)
	} else if sub, _ := idTok.Claim("sub").(string); sub != test_acntId {
		t.Error(sub)
		t.Fatal(test_acntId)
	} else if aud, _ := idTok.Claim("aud").(string); aud != test_taId {
		t.Error(aud)
		t.Fatal(test_taId)
	} else if email, _ := idTok.Claim("email").(string); email != test_email {
		t.Error(email)
		t.Fatal(test_email)
	} else if attr, _ := idTok.Claim("aaaaa").(string); attr != "bbbbb" {
		t.Error(attr)
		t.Fatal("bbbbb")
	}
}
