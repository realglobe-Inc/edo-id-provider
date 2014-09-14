package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"html"
	"net/http"
	"net/url"
	"time"
)

const cookieSessId = "SESSION_ID"

const (
	formUsrName = "user_name"
	formPasswd  = "password"

	formCliId   = "client_id"
	formRediUri = "redirect_uri"

	formSessLifetime = "session_lifetime"

	formCode   = "code"
	formCliSec = "client_secret"

	formAccTokenLifetime = "access_token_lifetime"

	formAccToken = "access_token"
	formAttr     = "attribute"
)

// /.
func routPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return erro.Wrap(err)
	}
	query := r.Form.Encode()
	if query != "" {
		query = "?" + query
	}

	sessIdCookie, err := r.Cookie(cookieSessId)
	if err != nil && err != http.ErrNoCookie {
		return erro.Wrap(err)
	} else if sessIdCookie == nil {
		// ログインページに飛ばす。
		w.Header().Set("Location", loginPagePath+query)
		w.WriteHeader(http.StatusFound)
		log.Debug("No session cookie.")
		return nil
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		// ログインページに飛ばす。
		w.Header().Set("Location", loginPagePath+query)
		w.WriteHeader(http.StatusFound)
		log.Debug("No valid session " + sessIdCookie.Value + ".")
		return nil
	}

	// 有効なセッションだった。
	log.Debug("Session " + sessIdCookie.Value + " is valid.")

	w.Header().Set("Location", setCookiePagePath+query)
	w.WriteHeader(http.StatusFound)
	log.Debug("Redirect to set cookie page.")
	return nil
}

// /login.
func loginPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	page := `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN">
<META HTTP-EQUIV="content-type" CONTENT="text/html; charset=utf-8">
<HTML>

  <HEAD>
    <TITLE>ログイン</TITLE>
  </HEAD>

  <BODY>
    <H1>ログインしてね。</H1>

    <P>
      <FORM METHOD="post" ACTION="` + beginSessPagePath + `">
        ユーザー名:<BR/><INPUT TYPE="text" NAME="` + formUsrName + `" SIZE="50" /><BR/>
        パスワード:<BR/><INPUT TYPE="password" NAME="` + formPasswd + `" SIZE="50" /><BR/>`

	if err := r.ParseForm(); err != nil {
		return erro.Wrap(err)
	}
	for key, vals := range r.Form {
		for _, val := range vals {
			page += `
        <INPUT TYPE="hidden" NAME="` + key + `" VALUE="` + html.EscapeString(val) + `" /> `
		}
	}

	page += `
        <INPUT TYPE="submit" VALUE="ログイン" /><BR/>
      </FORM>
    </P>
  </BODY>

</HTML>`

	w.Write([]byte(page))

	log.Debug("Responded login page.")
	return nil
}

// /logout.
func logoutPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	sessIdCookie, err := r.Cookie(cookieSessId)
	if err != nil && err != http.ErrNoCookie {
		return erro.Wrap(err)
	} else if sessIdCookie == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "no session was found.", nil))
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "session "+sessIdCookie.Value+" is invalid.", nil))
	}

	// 有効なセッションだった。
	log.Debug("Session " + sessIdCookie.Value + " is valid.")

	page := `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN">
<META HTTP-EQUIV="content-type" CONTENT="text/html; charset=utf-8">
<HTML>

  <HEAD>
    <TITLE>ログアウト</TITLE>
  </HEAD>

  <BODY>
    <H1>ログアウトするかい？</H1>

    <P>
      <FORM METHOD="post" ACTION="` + delCookiePagePath + `">
        <INPUT TYPE="submit" VALUE="ログアウト" /><BR/>`

	if err := r.ParseForm(); err != nil {
		return erro.Wrap(err)
	}
	for key, vals := range r.Form {
		for _, val := range vals {
			page += `
        <INPUT TYPE="hidden" NAME="` + key + `" VALUE="` + html.EscapeString(val) + `" /> `
		}
	}

	page += `
      </FORM>
    </P>
  </BODY>

</HTML>`

	w.Write([]byte(page))

	log.Debug("Responded logout page.")
	return nil
}

// /begin_session.
func beginSessionPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	usrName := r.FormValue(formUsrName)
	if usrName == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formUsrName+" parameter.", nil))
	}
	passwd := r.FormValue(formPasswd)
	if passwd == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formPasswd+" parameter.", nil))
	}
	r.Form.Del(formUsrName)
	r.Form.Del(formPasswd)

	usrUuid, err := sys.UserUuid(usrName)
	if err != nil {
		return erro.Wrap(err)
	} else if usrUuid == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "user "+usrName+" is not exist.", nil))
	}

	// ユーザー名が合ってた。
	log.Debug("User " + usrName + " found.")

	// TODO パスワードをハッシュ値にしとくとか。
	truePasswd, err := sys.UserPassword(usrUuid)
	if err != nil {
		return erro.Wrap(err)
	} else if passwd != truePasswd {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "wrong password for user "+usrName+".", nil))
	}

	// パスワードも合ってた。
	log.Debug("Right password for user " + usrName + ".")

	sess, err := sys.NewSession(usrUuid, sys.maxSessExpiDur) // 期限は /set_cookie で調整する。
	if err != nil {
		return erro.Wrap(err)
	}

	sessIdCookie := &http.Cookie{
		Name:   cookieSessId,
		Value:  sess.Id,
		MaxAge: int(sys.maxSessExpiDur.Seconds()), // 期限は /set_cookie で調整する。
	}
	w.Header().Set("Set-Cookie", sessIdCookie.String())

	query := r.Form.Encode()
	if query != "" {
		query = "?" + query
	}
	w.Header().Set("Location", setCookiePagePath+query)
	w.WriteHeader(http.StatusFound)

	log.Debug("Redirect to " + setCookiePagePath + ".")
	return nil
}

// /set_cookie.
func setCookiePage(sys *system, w http.ResponseWriter, r *http.Request) error {
	sessIdCookie, err := r.Cookie(cookieSessId)
	if err != nil && err != http.ErrNoCookie {
		return erro.Wrap(err)
	} else if sessIdCookie == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "no session was found.", nil))
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "session "+sessIdCookie.Value+" is invalid.", nil))
	}

	// 有効なセッションだった。
	log.Debug("Session " + sessIdCookie.Value + " is valid.")

	var expiDur time.Duration
	if expiDurStr := r.FormValue(formSessLifetime); expiDurStr != "" {
		var err error
		expiDur, err = time.ParseDuration(expiDurStr)
		if err != nil {
			return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "cannot parse "+formSessLifetime+" parameter "+expiDurStr+".", erro.Wrap(err)))
		}
	}
	if expiDur == 0 || expiDur > sys.maxSessExpiDur {
		expiDur = sys.maxSessExpiDur
	}

	sess.ExpiDate = time.Now().Add(expiDur)
	if err := sys.UpdateSession(sess); err != nil {
		return erro.Wrap(err)
	}

	// セッション期限が更新できた。
	log.Debug("Expiration date of session "+sessIdCookie.Value+" was updated to ", sess.ExpiDate, ".")

	newSessIdCookie := &http.Cookie{
		Name:   cookieSessId,
		Value:  sess.Id,
		MaxAge: int(expiDur.Seconds()),
	}
	w.Header().Set("Set-Cookie", newSessIdCookie.String())

	cliId := r.FormValue(formCliId)
	rediUri := r.FormValue(formRediUri)
	if cliId == "" || rediUri == "" {
		// ログイン済み（ログアウト）ページに飛ばす。
		if err := r.ParseForm(); err != nil {
			return erro.Wrap(err)
		}
		query := r.Form.Encode()
		if query != "" {
			query = "?" + query
		}
		w.Header().Set("Location", logoutPagePath+query)
		w.WriteHeader(http.StatusFound)
		log.Debug("Redirect to " + logoutPagePath + ".")
		return nil
	}

	// クライアントサービスのページにリダイレクトする必要あり。
	log.Debug("Need to redirect.")

	servUuid, err := sys.ServiceUuid(rediUri)
	if err != nil {
		return erro.Wrap(err)
	} else if servUuid != cliId {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "redirect uri "+rediUri+" does not belong to "+cliId+".", nil))
	}

	// クライアントサービスが登録されていて、リダイレクト先がクライアントサービスの管轄。
	log.Debug("Redirect destination " + rediUri + " belongs service " + cliId + ".")

	code, err := sys.NewCode(sess.UsrUuid, cliId)
	if err != nil {
		return erro.Wrap(err)
	}

	// code を発行。

	redi, err := url.Parse(rediUri)
	if err != nil {
		return erro.Wrap(err)
	}
	form := redi.Query()
	form.Set(formCode, code.Id)
	redi.RawQuery = form.Encode()
	rediUri = redi.String()
	w.Header().Set("Location", rediUri)
	w.WriteHeader(http.StatusFound)
	log.Debug("Redirect to " + rediUri + ".")
	return nil
}

// /delete_cookie.
func deleteCookiePage(sys *system, w http.ResponseWriter, r *http.Request) error {
	sessIdCookie, err := r.Cookie(cookieSessId)
	if err != nil && err != http.ErrNoCookie {
		return erro.Wrap(err)
	} else if sessIdCookie == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "no session was found.", nil))
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "session "+sessIdCookie.Value+" is invalid.", nil))
	}

	// 有効なセッションだった。
	log.Debug("Session " + sessIdCookie.Value + " is valid.")

	newSessIdCookie := &http.Cookie{
		Name:   cookieSessId,
		Value:  sess.Id,
		MaxAge: -1,
	}
	w.Header().Set("Set-Cookie", newSessIdCookie.String())

	cliId := r.FormValue(formCliId)
	rediUri := r.FormValue(formRediUri)
	if cliId == "" || rediUri == "" {
		// ログアウト済み（ログイン）ページに飛ばす。
		if err := r.ParseForm(); err != nil {
			return erro.Wrap(err)
		}
		query := r.Form.Encode()
		if query != "" {
			query = "?" + query
		}
		w.Header().Set("Location", loginPagePath+query)
		w.WriteHeader(http.StatusFound)
		log.Debug("Redirect to " + loginPagePath + ".")
		return nil
	}

	// クライアントサービスのページにリダイレクトする必要あり。
	log.Debug("Need to redirect.")

	servUuid, err := sys.ServiceUuid(rediUri)
	if err != nil {
		return erro.Wrap(err)
	} else if servUuid != cliId {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "redirect uri "+rediUri+" does not belong to "+cliId+".", nil))
	}

	// クライアントサービスが登録されていて、リダイレクト先がクライアントサービスの管轄。
	log.Debug("Redirect destination " + rediUri + " belongs  service " + cliId + ".")

	w.Header().Set("Location", rediUri)
	w.WriteHeader(http.StatusFound)
	log.Debug("Redirect to " + rediUri + ".")
	return nil
}

// /access_token.
func accessTokenPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	codeId := r.FormValue(formCode)
	if codeId == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formCode+" parameter.", nil))
	}
	cliId := r.FormValue(formCliId)
	if cliId == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formCliId+" parameter.", nil))
	}
	cliSec := r.FormValue(formCliSec)
	if cliSec == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formCliSec+" parameter.", nil))
	}
	var lifetime time.Duration
	if s := r.FormValue(formAccTokenLifetime); s != "" {
		var err error
		lifetime, err = time.ParseDuration(s)
		if err != nil {
			return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "cannot parse "+formAccTokenLifetime+" parameter "+s+".", erro.Wrap(err)))
		}
	}

	// パラメータはあった。

	code, err := sys.Code(codeId)
	if err != nil {
		return erro.Wrap(err)
	} else if code == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "code is invalid.", nil))
	}

	// code は有効だった。
	log.Debug("Code is valid.")

	if cliId != code.ServUuid {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "code is not bound to "+cliId+".", nil))
	}

	// 自称と発行先サービスが一致した。
	log.Debug("Code is bound to declared service " + cliId + ".")

	servKey, err := sys.ServiceKey(cliId)
	if err != nil {
		return erro.Wrap(err)
	} else if servKey == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "key of "+cliId+" is not exist.", nil))
	}

	// 公開鍵を取得できた。
	log.Debug("Key of " + cliId + " is exist.")

	// 署名検証。
	buff, err := base64.StdEncoding.DecodeString(cliSec)
	if err != nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "cannot parse "+formCliSec+" parameter.", erro.Wrap(err)))
	}

	if err := rsa.VerifyPKCS1v15(servKey, 0, []byte(codeId), buff); err != nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, formCliSec+" is invalid.", erro.Wrap(err)))
	}

	// 署名も正しかった。
	log.Debug(formCliSec + " is valid.")

	accToken, err := sys.NewAccessToken(code.UsrUuid, lifetime)
	if err != nil {
		return erro.Wrap(err)
	}

	var res struct {
		AccToken string `json:"access_token"`
	}
	res.AccToken = accToken.Id
	body, err := json.Marshal(&res)
	if err != nil {
		return erro.Wrap(err)
	}

	w.Header().Set("Content-Type", util.ContentTypeJson)
	w.Write(body)
	return nil
}

// /query.
func queryPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	accTokenId := r.FormValue(formAccToken)
	if accTokenId == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formAccToken+" parameter.", nil))
	}
	cliId := r.FormValue(formCliId)
	if cliId == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formCliId+" parameter.", nil))
	}
	cliSec := r.FormValue(formCliSec)
	if cliSec == "" {
		return erro.Wrap(util.NewHttpStatusError(http.StatusBadRequest, "no "+formCliSec+" parameter.", nil))
	}
	attrNames := r.Form[formAttr]
	if attrNames == nil {
		attrNames = []string{}
	}

	// パラメータはあった。

	accToken, err := sys.AccessToken(accTokenId)
	if err != nil {
		return erro.Wrap(err)
	} else if accToken == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "access token is invalid.", nil))
	}

	// code は有効だった。
	log.Debug("Access token is valid.")

	servKey, err := sys.ServiceKey(cliId)
	if err != nil {
		return erro.Wrap(err)
	} else if servKey == nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "key of "+cliId+" is not exist.", nil))
	}

	// 公開鍵を取得できた。
	log.Debug("Key of " + cliId + " is exist.")

	// 署名検証。
	buff, err := base64.StdEncoding.DecodeString(cliSec)
	if err != nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, "cannot parse "+formCliSec+" parameter.", erro.Wrap(err)))
	}

	if err := rsa.VerifyPKCS1v15(servKey, 0, []byte(accTokenId), buff); err != nil {
		return erro.Wrap(util.NewHttpStatusError(http.StatusForbidden, formCliSec+" is invalid.", erro.Wrap(err)))
	}

	// 署名も正しかった。
	log.Debug(formCliSec + " is valid.")

	var res struct {
		Usr map[string]interface{} `json:"user"`
	}
	res.Usr = map[string]interface{}{}

	for _, attrName := range attrNames {
		attr, err := sys.UserAttribute(accToken.UsrUuid, attrName)
		if err != nil {
			return erro.Wrap(err)
		}
		res.Usr[attrName] = attr
	}

	body, err := json.Marshal(&res)
	if err != nil {
		return erro.Wrap(err)
	}

	w.Header().Set("Content-Type", util.ContentTypeJson)
	w.Write(body)
	return nil
}
