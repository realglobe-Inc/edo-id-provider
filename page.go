package main

import (
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const cookieSessId = "SESSION_ID"

const (
	formUsrName = "user_name"
	formPasswd  = "password"
)

const (
	queryCliId   = "client_id"
	queryRediUri = "redirect_uri"

	querySessLifetime = "session_lifetime"
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
      <FORM METHOD="post" ACTION="` + beginSessPagePath + `" METHOD="POST">
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

	log.Debug("Response login page.")
	return nil
}

// /logout.
func logoutPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	sessIdCookie, err := r.Cookie(cookieSessId)
	if err != nil && err != http.ErrNoCookie {
		return erro.Wrap(err)
	} else if sessIdCookie == nil {
		return erro.Wrap(newInvalidSession())
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(newInvalidSession())
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
      <FORM METHOD="post" ACTION="` + delCookiePagePath + `" METHOD="POST">
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

	log.Debug("Response logout page.")
	return nil
}

// /begin_session.
func beginSessionPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	usrName := r.FormValue(formUsrName)
	if usrName == "" {
		return erro.Wrap(newInvalidRequest("no " + formUsrName + " parameter."))
	}
	passwd := r.FormValue(formPasswd)
	if passwd == "" {
		return erro.Wrap(newInvalidRequest("no " + formPasswd + " parameter."))
	}
	r.Form.Del(formUsrName)
	r.Form.Del(formPasswd)

	var lifetime time.Duration
	if s := r.FormValue(querySessLifetime); s != "" {
		var err error
		lifetime, err = time.ParseDuration(s)
		if err != nil {
			return erro.Wrap(err)
		}
	} else {
		lifetime = sys.maxSessExpiDur
	}

	usrUuid, err := sys.UserUuid(usrName)
	if err != nil {
		return erro.Wrap(err)
	} else if usrUuid == "" {
		return erro.Wrap(newUserNotFound(usrName))
	}

	// ユーザー名が合ってた。
	log.Debug("User " + usrName + " found.")

	// TODO パスワードをハッシュ値にしとくとか。
	truePasswd, err := sys.UserPassword(usrUuid)
	if err != nil {
		return erro.Wrap(err)
	} else if passwd != truePasswd {
		return erro.Wrap(newInvalidPassword(usrName))
	}

	// パスワードも合ってた。
	log.Debug("User " + usrName + " password is correct.")

	sess, err := sys.NewSession(usrUuid, lifetime)
	if err != nil {
		return erro.Wrap(err)
	}

	sessIdCookie := &http.Cookie{
		Name:   cookieSessId,
		Value:  sess.Id,
		MaxAge: int(sess.ExpiDate.Sub(time.Now()).Seconds()),
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
		return erro.Wrap(newInvalidSession())
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(newInvalidSession())
	}

	// 有効なセッションだった。
	log.Debug("Session " + sessIdCookie.Value + " is valid.")

	newSessIdCookie := &http.Cookie{
		Name:   cookieSessId,
		Value:  sess.Id,
		MaxAge: sys.cookieMaxAge,
	}
	w.Header().Set("Set-Cookie", newSessIdCookie.String())

	cliId, rediUri := r.FormValue(queryCliId), r.FormValue(queryRediUri)
	if cliId == "" && rediUri == "" {
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
		return erro.Wrap(newInvalidRedirectUri(cliId, rediUri))
	}

	// クライアントサービスが登録されていて、リダイレクト先がクライアントサービスの管轄。
	log.Debug("Redirect destination " + rediUri + " belongs service " + cliId + ".")

	code, err := sys.NewCode(cliId)
	if err != nil {
		return erro.Wrap(err)
	}

	// code を発行。

	var query string
	if strings.Index(rediUri, "?") < 0 {
		query = "?code=" + url.QueryEscape(code.Id)
	} else {
		query = "&code=" + url.QueryEscape(code.Id)
	}
	w.Header().Set("Location", rediUri+query)
	w.WriteHeader(http.StatusFound)
	log.Debug("Redirect to " + rediUri + ".")
	return nil
}

// /del_cookie.
func delCookiePage(sys *system, w http.ResponseWriter, r *http.Request) error {
	sessIdCookie, err := r.Cookie(cookieSessId)
	if err != nil && err != http.ErrNoCookie {
		return erro.Wrap(err)
	} else if sessIdCookie == nil {
		return erro.Wrap(newInvalidSession())
	}

	// cookie にセッションがあった。
	log.Debug("Session " + sessIdCookie.Value + " is in cookie.")

	sess, err := sys.Session(sessIdCookie.Value)
	if err != nil {
		return erro.Wrap(err)
	} else if sess == nil {
		return erro.Wrap(newInvalidSession())
	}

	// 有効なセッションだった。
	log.Debug("Session " + sessIdCookie.Value + " is valid.")

	newSessIdCookie := &http.Cookie{
		Name:   cookieSessId,
		Value:  sess.Id,
		MaxAge: -1,
	}
	w.Header().Set("Set-Cookie", newSessIdCookie.String())

	cliId, rediUri := r.FormValue(queryCliId), r.FormValue(queryRediUri)
	if cliId == "" && rediUri == "" {
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
		return erro.Wrap(newInvalidRedirectUri(cliId, rediUri))
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
	panic("not yet implemented.")
}

// /query.
func queryPage(sys *system, w http.ResponseWriter, r *http.Request) error {
	panic("not yet implemented.")
}
