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
	"github.com/realglobe-Inc/edo-id-provider/claims"
	"github.com/realglobe-Inc/edo-id-provider/database/consent"
	"github.com/realglobe-Inc/edo-lib/strset/strsetutil"
	"reflect"
	"testing"
)

func TestProvidedScopes(t *testing.T) {
	scopCons := consent.NewConsent("openid", "email")
	reqScops := strsetutil.New("openid", "email")
	ans := strsetutil.New("openid", "email")

	if scops, err := ProvidedScopes(scopCons, reqScops); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(scops, ans) {
		t.Error(ans)
		t.Fatal(scops)
	}
}

func TestProvidedScopesShrink(t *testing.T) {
	scopCons := consent.NewConsent("openid")
	reqScops := strsetutil.New("openid", "email")
	ans := strsetutil.New("openid")

	if scops, err := ProvidedScopes(scopCons, reqScops); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(scops, ans) {
		t.Error(ans)
		t.Fatal(scops)
	}
}

func TestProvidedScopesDenied(t *testing.T) {
	scopCons := consent.NewConsent("")
	reqScops := strsetutil.New("openid", "email")

	if _, err := ProvidedScopes(scopCons, reqScops); err == nil {
		t.Fatal("no error")
	}
}

func TestProvidedAttributes(t *testing.T) {
	scopCons := consent.NewConsent("openid")
	attrCons := consent.NewConsent("email")
	reqClms := claims.Claims{
		"email": claims.New(false, nil, nil, ""),
	}
	ans := strsetutil.New("sub", "email")

	if attrs, err := ProvidedAttributes(scopCons, attrCons, nil, reqClms); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(attrs, ans) {
		t.Error(ans)
		t.Fatal(attrs)
	}
}

func TestProvidedAttributesShrink(t *testing.T) {
	scopCons := consent.NewConsent("openid")
	attrCons := consent.NewConsent()
	reqClms := claims.Claims{
		"email": claims.New(false, nil, nil, ""),
	}
	ans := strsetutil.New("sub")

	if attrs, err := ProvidedAttributes(scopCons, attrCons, nil, reqClms); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(attrs, ans) {
		t.Error(ans)
		t.Fatal(attrs)
	}
}

func TestProvidedAttributesDenied(t *testing.T) {
	scopCons := consent.NewConsent("openid")
	attrCons := consent.NewConsent()
	reqClms := claims.Claims{
		"email": claims.New(true, nil, nil, ""),
	}

	if _, err := ProvidedAttributes(scopCons, attrCons, nil, reqClms); err == nil {
		t.Fatal(err)
	}
}

func TestProvidedAttributesByScope(t *testing.T) {
	scopCons := consent.NewConsent("openid")
	attrCons := consent.NewConsent("email")
	scops := strsetutil.New("email")
	ans := strsetutil.New("sub", "email", "email_verified")

	if attrs, err := ProvidedAttributes(scopCons, attrCons, scops, nil); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(attrs, ans) {
		t.Error(ans)
		t.Fatal(attrs)
	}
}
