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

import ()

// コンパイル時に打ち間違いを検知するため。それ以上ではない。

// HTTP の URL クエリや application/json, application/x-www-form-urlencoded で使うパラメータ名。
const (
	// OAuth と OpenID Connect で定義されているパラメータ。
	formAccess_token          = "access_token"
	formClaim                 = "claim"
	formClaims                = "claims"
	formClient_assertion      = "client_assertion"
	formClient_assertion_type = "client_assertion_type"
	formClient_id             = "client_id"
	formClient_secret         = "client_secret"
	formCode                  = "code"
	formDisplay               = "display"
	formError                 = "error"
	formError_description     = "error_description"
	formExpires_in            = "expires_in"
	formGrant_type            = "grant_type"
	formId_token              = "id_token"
	formMax_age               = "max_age"
	formNonce                 = "nonce"
	formPrompt                = "prompt"
	formRedirect_uri          = "redirect_uri"
	formRefresh_token         = "refresh_token"
	formRequest               = "request"
	formRequest_uri           = "request_uri"
	formResponse_type         = "response_type"
	formScope                 = "scope"
	formState                 = "state"
	formToken_type            = "token_type"
	formUi_locales            = "ui_locales"

	// 独自。
	formAllowed_claims  = "allowed_claims"
	formAllowed_scope   = "allowed_scope"
	formDenied_claims   = "denied_claims"
	formDenied_scope    = "denied_scope"
	formIssuer          = "issuer"
	formLocale          = "locale"
	formLocales         = "locales"
	formMessage         = "message"
	formOptional_claims = "optional_claims"
	formPass_type       = "pass_type"
	formPassword        = "password"
	formTicket          = "ticket"
	formUsername        = "username"
	formUsernames       = "usernames"
)

// Cookie 中にセッション識別子。
const sessLabel = "Id-Provider"

// scope の値。
const (
	scopOpenid         = "openid"
	scopOffline_access = "offline_access"
)

// response_type の値。
const (
	respTypeCode     = "code"
	respTypeId_token = "id_token"
)

// prompt の値。
const (
	prmptConsent        = "consent"
	prmptLogin          = "login"
	prmptNone           = "none"
	prmptSelect_account = "select_account"
)

// client_assertion_type の値。
const (
	taAssTypeJwt = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

// クレーム名。
const (
	clmAt_hash   = "at_hash"
	clmAud       = "aud"
	clmAuth_time = "auth_time"
	clmC_hash    = "at_hash"
	clmExp       = "exp"
	clmIat       = "iat"
	clmIss       = "iss"
	clmJti       = "jti"
	clmNonce     = "nonce"
	clmSub       = "sub"
)

// JWT のヘッダパラメータ名。
const (
	jwtAlg = "alg"
	jwtKid = "kid"
)

// JWT のヘッダ alg の値。
const (
	algNone = "none"
)

// grant_type の値。
const (
	grntTypeAuthorization_code = "authorization_code"
)

// token_type の値。
const (
	tokTypeBearer = "Bearer"
)

// HTTP ヘッダ名。
const (
	headAuthorization = "Authorization"
)

// HTTP の Authorization の方式名。
const (
	scmBearer = "Bearer"
)
