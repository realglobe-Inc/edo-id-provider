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

const (
	// アンダースコア。
	tagAllowed_claims  = "allowed_claims"
	tagAllowed_scope   = "allowed_scope"
	tagAuth_time       = "auth_time"
	tagC_hash          = "c_hash"
	tagClaims          = "claims"
	tagClient_id       = "client_id"
	tagCode            = "code"
	tagConsent         = "consent"
	tagDenied_claims   = "denied_claims"
	tagDenied_scope    = "denied_scope"
	tagDisplay         = "display"
	tagExpires_in      = "expires_in"
	tagId_token        = "id_token"
	tagIssuer          = "issuer"
	tagLocale          = "locale"
	tagLocales         = "locales"
	tagLogin           = "login"
	tagMessage         = "message"
	tagNonce           = "nonce"
	tagNone            = "none"
	tagOpenid          = "openid"
	tagOptional_claims = "optional_claims"
	tagPass_type       = "pass_type"
	tagPassword        = "password"
	tagRequest         = "request"
	tagRequest_uri     = "request_uri"
	tagScope           = "scope"
	tagSelect_account  = "select_account"
	tagState           = "state"
	tagTicket          = "ticket"
	tagUsername        = "username"
	tagUsernames       = "usernames"

	// ハイフン。
	tagNo_cache = "no-cache"
	tagNo_store = "no-store"

	// 頭大文字、ハイフン。
	tagCache_control = "Cache-Control"
	tagPragma        = "Pragma"
)
