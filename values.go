package main

import ()

// HTTP の URL クエリや application/json, application/x-www-form-urlencoded で使うパラメータ名。
const (
	// OAuth と OpenID Connect で定義されているパラメータ。
	formScop      = "scope"
	formTaId      = "client_id"
	formTaScrt    = "client_secret"
	formPrmpt     = "prompt"
	formRediUri   = "redirect_uri"
	formRespType  = "response_type"
	formStat      = "state"
	formNonc      = "nonce"
	formCod       = "code"
	formGrntType  = "grant_type"
	formTaAssType = "client_assertion_type"
	formTaAss     = "client_assertion"
	formTokId     = "access_token"
	formTokType   = "token_type"
	formExpi      = "expires_in"
	formRefTok    = "refresh_token"
	formIdTok     = "id_token"
	formErr       = "error"
	formErrDesc   = "error_description"
	formClm       = "claim"
	formDisp      = "display"
	formMaxAge    = "max_age"
	formUiLocs    = "ui_locales"

	// 独自。
	formAccName   = "username"
	formPasswd    = "password"
	formAccNames  = "usernames"
	formHint      = "hint"
	formLoginTic  = "ticket"
	formTaNam     = "client_name"
	formConsTic   = "ticket"
	formConsScops = "consented_scope"
	formConsClms  = "consented_claim"
	formDenyScops = "denied_scope"
	formDenyClms  = "denied_claim"
	formSelTic    = "ticket"
	formLocs      = "locales"
)

// Cookie 中にセッション識別子。
const cookSess = "X-Edo-Idp-Session"

// scope の値。
const (
	scopOpId = "openid"
)

// response_type の値。
const (
	respTypeCod = "code"
)

// prompt の値。
const (
	prmptSelAcc = "select_account"
	prmptLogin  = "login"
	prmptCons   = "consent"
	prmptNone   = "none"
)

// client_assertion_type の値。
const (
	taAssTypeJwt = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
)

// クレーム名。
const (
	clmIss     = "iss"
	clmSub     = "sub"
	clmAud     = "aud"
	clmJti     = "jti"
	clmExp     = "exp"
	clmIat     = "iat"
	clmAuthTim = "auth_time"
	clmNonc    = "nonce"
	clmAtHash  = "at_hash"

	// プライベートクレーム。
	clmCod = "code"
	clmTok = "access_token"
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
	grntTypeCod = "authorization_code"
)

// token_type の値。
const (
	tokTypeBear = "Bearer"
)

// HTTP ヘッダ名。
const (
	headAuth = "Authorization"
)

// HTTP の Authorization の方式名。
const (
	scBear = "Bearer"
)

// error の値。
const (
	errInvReq         = "invalid_request"
	errAccDeny        = "access_dnied"
	errUnsuppRespType = "unsupported_response_type"
	errInvScop        = "invalid_scope"
	errServErr        = "server_error"
	errInteractReq    = "interaction_required"
	errLoginReq       = "login_required"
	errAccSelReq      = "account_selection_required"
	errConsReq        = "consent_required"
	errReqNotSupp     = "request_not_supported"
	errReqUriNotSupp  = "request_uri_not_supported"
	errRegNotSupp     = "registration_not_supported"
	errUnsuppGrntType = "unsupported_grant_type"
	errInvGrnt        = "invalid_grant"
	errInvTa          = "invalid_client"
	// OpenID Connect の仕様ではサンプルとしてしか登場しない。
	errInvTok = "invalid_token"
)

// URL パス。
const (
	authPath   = "/auth"
	loginPath  = "/auth/login"
	selPath    = "/auth/select"
	consPath   = "/auth/consent"
	tokPath    = "/token"
	accInfPath = "/userinfo"
)

// HTML ファイル名。
const (
	selHtml   = "select.html"
	loginHtml = "login.html"
	consHtml  = "consent.html"
)
