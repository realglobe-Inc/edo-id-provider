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
const (
	// HTTP の URL クエリや application/json, application/x-www-form-urlencoded で使うパラメータ名。
	// OAuth と OpenID Connect で定義されているパラメータ。
	tagAccess_token          = "access_token"
	tagClaim                 = "claim"
	tagClaims                = "claims"
	tagClient_assertion      = "client_assertion"
	tagClient_assertion_type = "client_assertion_type"
	tagClient_id             = "client_id"
	tagClient_secret         = "client_secret"
	tagCode                  = "code"
	tagDisplay               = "display"
	tagError                 = "error"
	tagError_description     = "error_description"
	tagExpires_in            = "expires_in"
	tagGrant_type            = "grant_type"
	tagId_token              = "id_token"
	tagMax_age               = "max_age"
	tagNonce                 = "nonce"
	tagPrompt                = "prompt"
	tagRedirect_uri          = "redirect_uri"
	tagRefresh_token         = "refresh_token"
	tagRequest               = "request"
	tagRequest_uri           = "request_uri"
	tagResponse_type         = "response_type"
	tagScope                 = "scope"
	tagState                 = "state"
	tagToken_type            = "token_type"
	tagUi_locales            = "ui_locales"

	// 独自。
	tagAllowed_claims  = "allowed_claims"
	tagAllowed_scope   = "allowed_scope"
	tagDenied_claims   = "denied_claims"
	tagDenied_scope    = "denied_scope"
	tagIssuer          = "issuer"
	tagLocale          = "locale"
	tagLocales         = "locales"
	tagMessage         = "message"
	tagOptional_claims = "optional_claims"
	tagPass_type       = "pass_type"
	tagPassword        = "password"
	tagTicket          = "ticket"
	tagUsername        = "username"
	tagUsernames       = "usernames"

	// scope の値。
	tagOpenid         = "openid"
	tagOffline_access = "offline_access"

	// response_type の値。
	//tagCode     = "code"
	//tagId_token = "id_token"

	// prompt の値。
	tagConsent        = "consent"
	tagLogin          = "login"
	tagNone           = "none"
	tagSelect_account = "select_account"

	// クレーム名。
	tagAt_hash   = "at_hash"
	tagAud       = "aud"
	tagAuth_time = "auth_time"
	tagC_hash    = "at_hash"
	tagExp       = "exp"
	tagIat       = "iat"
	tagIss       = "iss"
	tagJti       = "jti"
	//tagNonce     = "nonce"
	tagSub = "sub"

	// JWT のヘッダ名。
	tagAlg = "alg"
	tagKid = "kid"

	// JWT の alg ヘッダの値。
	//tagNone = "none"

	// grant_type の値。
	tagAuthorization_code = "authorization_code"

	// token_type の値。
	tagBearer = "Bearer"

	// HTTP メソッド。
	tagPost = "POST"

	// HTTP ヘッダ名。
	tagAuthorization  = "Authorization"
	tagCache_control  = "Cache-Control"
	tagContent_length = "Content-Length"
	tagContent_type   = "Content-Type"
	tagPragma         = "Pragma"

	// HTTP ヘッダ値。
	//tagBearer = "Bearer"
	tagNo_store = "no-store"
	tagNo_cache = "no-cache"
)

const (
	// client_assertion_type の値。
	cliAssTypeJwt_bearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)
