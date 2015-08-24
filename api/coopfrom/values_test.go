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

package coopfrom

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/realglobe-Inc/edo-id-provider/database/account"
	hashutil "github.com/realglobe-Inc/edo-id-provider/hash"
	idpdb "github.com/realglobe-Inc/edo-idp-selector/database/idp"
	tadb "github.com/realglobe-Inc/edo-idp-selector/database/ta"
	"github.com/realglobe-Inc/edo-lib/jwk"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"github.com/realglobe-Inc/edo-lib/jwt/audience"
	"github.com/realglobe-Inc/go-lib/erro"
)

const (
	test_sigAlg = "ES256"
	test_hAlg   = "SHA256"

	test_tokId = "ZkTPOdBdh_bS2PqWnb1r8A3DqeKGCC"

	test_acntTag   = "main-user"
	test_acntId    = "EYClXo4mQKwSgPel"
	test_acntEmail = "tester@example.org"

	test_subAcnt1Tag   = "sub-user1"
	test_subAcnt1Id    = "U7pdvT8dYbBFWXdc"
	test_subAcnt1Email = "subtester1@example.org"

	test_subAcnt2Tag   = "sub-user2"
	test_subAcnt2Id    = "lgmxuHfXfSTB-1js"
	test_subAcnt2Email = "subtester2@example.org"

	test_frTaSigAlg = "ES384"
	test_jti        = "R-seIeMPBly4xPAh"
)

var (
	test_idpKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "lpHYO1qpjU95B2sThPR2-1jv44axgaEDkQtcKNE-oZs",
		"y":   "soy5O11SFFFeYdhQVodXlYPIpeo0pCS69IxiVPPf0Tk",
		"d":   "3BhkCluOkm8d8gvaPD5FDG2zeEw2JKf3D5LwN-mYmsw",
	})

	test_acntAttrs = map[string]interface{}{
		"email": test_acntEmail,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}
	test_subAcnt1Attrs = map[string]interface{}{
		"email": test_subAcnt1Email,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}
	test_subAcnt2Attrs = map[string]interface{}{
		"email": test_subAcnt2Email,
		"pds": map[string]interface{}{
			"type": "single",
			"uri":  "https://pds.example.org",
		},
	}

	test_frTaKey, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-384",
		"x":   "HlrMhzZww_AkmHV-2gDR5n7t75673UClnC7V2GewWva_sg-4GSUguFalVgwnK0tQ",
		"y":   "fxS48Fy50SZFZ-RAQRWUZXZgRSWwiKVkqPTd6gypfpQNkXSwE69BXYIAQcfaLcf2",
		"d":   "Gp-7eC0G7PjGzKoiAmTQ1iLsLU3AEy3h-bKFWSZOanXqSWI6wqJVPEUsatNYBJoG",
	})
	test_frTa = tadb.New("https://from.example.org", nil, nil, []jwk.Key{test_frTaKey}, false, "")
	test_toTa = tadb.New("https://to.example.org", nil, nil, nil, false, "")

	test_idp2Key, _ = jwk.FromMap(map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   "vQ3EYqVi30Zd4NF0hbKdHIMZAngSrhwa3mxx74zXkDc",
		"y":   "OwPvhvTL0SlgB7SpucwBOyjbbY0V8M1-dS6FwkMPGD8",
		"d":   "Y4YXo4D_B5FMj_5oXizubBDWRWETRpWr8jX969odblA",
	})
	test_idp2 = idpdb.New("https://idp.example.org", nil, "", "", "", "", "", []jwk.Key{test_idp2Key})
)

func newTestMainAccount() account.Element {
	return account.New(test_acntId, "", nil, clone(test_acntAttrs))
}

func newTestSubAccount1() account.Element {
	return account.New(test_subAcnt1Id, "", nil, clone(test_subAcnt1Attrs))
}

func newTestSubAccount2() account.Element {
	return account.New(test_subAcnt2Id, "", nil, clone(test_subAcnt2Attrs))
}

// 1 段目だけのコピー。
func clone(m map[string]interface{}) map[string]interface{} {
	m2 := map[string]interface{}{}
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func calcTestSubAccount2HashValue(idp string) string {
	return calcTestAccountHashValue(idp, test_subAcnt2Id)
}

func calcTestAccountHashValue(idp, id string) string {
	hGen := hashutil.Generator(test_hAlg)
	if !hGen.Available() {
		panic("unsupported hash algorithm " + test_hAlg)
	}
	hFun := hGen.New()
	hFun.Write([]byte(idp))
	hFun.Write([]byte{0})
	hFun.Write([]byte(id))
	hVal := hFun.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(hVal[:len(hVal)/2])
}

func newTestSingleRequest(aud string) (*http.Request, error) {
	return newTestSingleRequestWithParams(aud, nil, nil)
}

func newTestSingleRequestWithParams(aud string, params, assParams map[string]interface{}) (*http.Request, error) {
	m := map[string]interface{}{
		"response_type":         "code_token",
		"from_client":           test_frTa.Id(),
		"to_client":             test_toTa.Id(),
		"grant_type":            "access_token",
		"access_token":          test_tokId,
		"user_tag":              test_acntTag,
		"users":                 map[string]string{test_subAcnt1Tag: test_subAcnt1Id},
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}
	{
		ass := jwt.New()
		now := time.Now()
		ass.SetHeader("alg", test_frTaSigAlg)
		ass.SetClaim("iss", test_frTa.Id())
		ass.SetClaim("sub", test_frTa.Id())
		ass.SetClaim("aud", audience.New(aud))
		ass.SetClaim("jti", test_jti)
		ass.SetClaim("exp", now.Add(time.Minute).Unix())
		ass.SetClaim("iat", now.Unix())
		for k, v := range assParams {
			ass.SetClaim(k, v)
		}
		if err := ass.Sign(test_frTa.Keys()); err != nil {
			return nil, erro.Wrap(err)
		}
		buff, err := ass.Encode()
		if err != nil {
			return nil, erro.Wrap(err)
		}
		m["client_assertion"] = string(buff)
	}
	for k, v := range params {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = v
		}
	}
	body, err := json.Marshal(m)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r, err := http.NewRequest("POST", aud, bytes.NewReader(body))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("Content-Type", "application/json")
	return r, nil
}

func newTestMainRequest(aud, subIdp string) (*http.Request, error) {
	return newTestMainRequestWithParams(aud, subIdp, nil, nil)
}

func newTestMainRequestWithParams(aud, subIdp string, params, assParams map[string]interface{}) (*http.Request, error) {
	m := map[string]interface{}{
		"response_type":         "code_token referral",
		"from_client":           test_frTa.Id(),
		"to_client":             test_toTa.Id(),
		"grant_type":            "access_token",
		"access_token":          test_tokId,
		"user_tag":              test_acntTag,
		"users":                 map[string]string{test_subAcnt1Tag: test_subAcnt1Id},
		"related_users":         map[string]string{test_subAcnt2Tag: calcTestSubAccount2HashValue(subIdp)},
		"hash_alg":              test_hAlg,
		"related_issuers":       []string{subIdp},
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}
	{
		ass := jwt.New()
		now := time.Now()
		ass.SetHeader("alg", test_frTaSigAlg)
		ass.SetClaim("iss", test_frTa.Id())
		ass.SetClaim("sub", test_frTa.Id())
		ass.SetClaim("aud", audience.New(aud))
		ass.SetClaim("jti", test_jti)
		ass.SetClaim("exp", now.Add(time.Minute).Unix())
		ass.SetClaim("iat", now.Unix())
		for k, v := range assParams {
			ass.SetClaim(k, v)
		}
		if err := ass.Sign(test_frTa.Keys()); err != nil {
			return nil, erro.Wrap(err)
		}
		buff, err := ass.Encode()
		if err != nil {
			return nil, erro.Wrap(err)
		}
		m["client_assertion"] = string(buff)
	}
	for k, v := range params {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = v
		}
	}
	body, err := json.Marshal(m)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r, err := http.NewRequest("POST", aud, bytes.NewReader(body))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	r.Header.Set("Content-Type", "application/json")
	return r, nil
}

func newTestSubRequest(idp, aud string) (r *http.Request, refHash []byte, err error) {
	return newTestSubRequestWithParams(idp, aud, nil, nil, nil)
}

func newTestSubRequestWithParams(idp, aud string, params, refParams, assParams map[string]interface{}) (r *http.Request, refHash []byte, err error) {
	ref := jwt.New()
	ref.SetHeader("alg", test_sigAlg)
	ref.SetClaim("iss", test_idp2.Id())
	ref.SetClaim("sub", test_frTa.Id())
	ref.SetClaim("aud", audience.New(idp))
	ref.SetClaim("exp", time.Now().Add(time.Minute).Unix())
	ref.SetClaim("jti", test_jti)
	ref.SetClaim("to_client", test_toTa.Id())
	ref.SetClaim("related_users", map[string]string{test_subAcnt2Tag: calcTestSubAccount2HashValue(idp)})
	ref.SetClaim("hash_alg", test_hAlg)
	for k, v := range refParams {
		ref.SetClaim(k, v)
	}
	if err := ref.Sign([]jwk.Key{test_idp2Key}); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	refBuff, err := ref.Encode()
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	m := map[string]interface{}{
		"response_type":         "code_token",
		"grant_type":            "referral",
		"referral":              string(refBuff),
		"users":                 map[string]string{test_subAcnt2Tag: test_subAcnt2Id},
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	}
	{
		ass := jwt.New()
		now := time.Now()
		ass.SetHeader("alg", test_frTaSigAlg)
		ass.SetClaim("iss", test_frTa.Id())
		ass.SetClaim("sub", test_frTa.Id())
		ass.SetClaim("aud", audience.New(aud))
		ass.SetClaim("jti", test_jti)
		ass.SetClaim("exp", now.Add(time.Minute).Unix())
		ass.SetClaim("iat", now.Unix())
		for k, v := range assParams {
			ref.SetClaim(k, v)
		}
		if err := ass.Sign(test_frTa.Keys()); err != nil {
			return nil, nil, erro.Wrap(err)
		}
		buff, err := ass.Encode()
		if err != nil {
			return nil, nil, erro.Wrap(err)
		}
		m["client_assertion"] = string(buff)
	}
	for k, v := range params {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = v
		}
	}
	body, err := json.Marshal(m)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	r, err = http.NewRequest("POST", aud, bytes.NewReader(body))
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}
	r.Header.Set("Content-Type", "application/json")

	hGen := hashutil.Generator(test_hAlg)
	if hGen == 0 {
		return nil, nil, erro.New("unsupported hash algorithm " + test_hAlg)
	}
	hFun := hGen.New()
	hFun.Write(refBuff)
	hVal := hFun.Sum(nil)

	return r, hVal[:len(hVal)/2], nil
}
