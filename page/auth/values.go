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

package auth

import ()

const (
	// アンダースコア。
	// tagAccess_token          = "access_token"
	// tagAlg                   = "alg"
	tagAllowed_claims = "allowed_claims"
	tagAllowed_scope  = "allowed_scope"
	// tagAt_hash               = "at_hash"
	// tagAud                   = "aud"
	tagAuth_time = "auth_time"
	// tagAuthorization_code    = "authorization_code"
	tagC_hash = "at_hash"
	// tagClaim                 = "claim"
	tagClaims = "claims"
	// tagClient_assertion      = "client_assertion"
	// tagClient_assertion_type = "client_assertion_type"
	tagClient_id = "client_id"
	// tagClient_secret         = "client_secret"
	tagCode = "code"
	// tagCode_token            = "code_token"
	tagConsent       = "consent"
	tagDenied_claims = "denied_claims"
	tagDenied_scope  = "denied_scope"
	tagDisplay       = "display"
	// tagError                 = "error"
	// tagError_description     = "error_description"
	// tagExp                   = "exp"
	tagExpires_in = "expires_in"
	// tagGrant_type            = "grant_type"
	// tagHash_alg              = "hash_alg"
	// tagIat                   = "iat"
	tagId_token = "id_token"
	// tagIss                   = "iss"
	tagIssuer = "issuer"
	// tagJti                   = "jti"
	// tagKid                   = "kid"
	tagLocale  = "locale"
	tagLocales = "locales"
	tagLogin   = "login"
	// tagMax_age               = "max_age"
	tagMessage = "message"
	tagNonce   = "nonce"
	tagNone    = "none"
	// tagOffline_access        = "offline_access"
	tagOpenid          = "openid"
	tagOptional_claims = "optional_claims"
	tagPass_type       = "pass_type"
	tagPassword        = "password"
	// tagPrompt                = "prompt"
	// tagRedirect_uri          = "redirect_uri"
	// tagRef_hash              = "ref_hash"
	// tagReferral              = "referral"
	// tagRefresh_token         = "refresh_token"
	// tagRelated_users         = "related_users"
	tagRequest     = "request"
	tagRequest_uri = "request_uri"
	// tagResponse_type         = "response_type"
	tagScope          = "scope"
	tagSelect_account = "select_account"
	tagState          = "state"
	// tagSub                   = "sub"
	tagTicket = "ticket"
	// tagTo_client             = "to_client"
	// tagToken_type            = "token_type"
	// tagUi_locales            = "ui_locales"
	// tagUser_tag              = "user_tag"
	// tagUser_tags             = "user_tags"
	tagUsername  = "username"
	tagUsernames = "usernames"

	// ハイフン。
	tagNo_cache = "no-cache"
	tagNo_store = "no-store"

	// 頭大文字、ハイフン。
	// tagAuthorization  = "Authorization"
	// tagBearer         = "Bearer"
	tagCache_control = "Cache-Control"
	// tagContent_length = "Content-Length"
	// tagContent_type   = "Content-Type"
	tagPragma = "Pragma"

	// 大文字。
	// tagPost = "POST"
)

const (
// client_assertion_type の値。
// cliAssTypeJwt_bearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

// Content-Type の値。
// contTypeHtml = "text/html"
// contTypeJson = "application/json"
)
