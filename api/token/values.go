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

package token

const (
	// アンダースコア。
	tagAccess_token          = "access_token"
	tagAt_hash               = "at_hash"
	tagAuth_time             = "auth_time"
	tagAuthorization_code    = "authorization_code"
	tagClient_assertion      = "client_assertion"
	tagClient_assertion_type = "client_assertion_type"
	tagClient_id             = "client_id"
	tagClient_secret         = "client_secret"
	tagCode                  = "code"
	tagExpires_in            = "expires_in"
	tagGrant_type            = "grant_type"
	tagId_token              = "id_token"
	tagNonce                 = "nonce"
	tagRedirect_uri          = "redirect_uri"
	tagRefresh_token         = "refresh_token"
	tagScope                 = "scope"
	tagToken_type            = "token_type"

	// 頭大文字、ハイフン。
	tagAuthorization = "Authorization"
	tagBearer        = "Bearer"

	// 大文字。
	tagPost = "POST"
)

const (
	// client_assertion_type の値。
	cliAssTypeJwt_bearer = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)
