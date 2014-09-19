edo-id-provider
===

ログインサーバー。

### /

cookie に有効な SESSION_ID が設定されていれば、
URL パラメータを維持したまま /set_cookie にリダイレクト。
設定されていなければ、URL パラメータを維持したまま /login にリダイレクト。

### /login

ログイン画面を表示する。
ログインボタンを押すと、
/begin_session に元の URL パラメータをつけて、
ユーザー名とパスワードを POST する。

### /begin_session

user_name と password が付随していなければ 400 Bad Request。

ユーザー名とパスワードが正しければ、
cookie に SESSION_ID を設定して、
URL パラメータを維持したまま /set_cookie にリダイレクト。
正しくなければ、403 Forbidden。

(TODO) OpenID Connect 的にはエラーは URL パラメータで返すらしい。

### /set_cookie?client_id={client_id}&redirect_uri={redirect_uri}

cookie に有効な SESSION_ID が設定されていて、
client_id が登録されているサービスの UUID で、
redirect_uri がそのサービスの提供する URI 以下ならば、
SESSION_ID を更新して、
code をつけて redirect_uri にリダイレクト。
そうでなければ、403 Forbidden。

### /set_cookie

cookie に有効な SESSION_ID が設定されているならば、
SESSION_ID を更新して、
URL パラメータを維持したまま /logout にリダイレクト。
設定されていなければ、403 Forbidden。

### /logout

cookie に有効な SESSION_ID が設定されていれば、ログアウト画面を表示する。
設定されていなければ、403 Forbidden。

(TODO) /login に戻す手もある。ただ、無限ループバグを作る可能性あり。

ログアウトボタンを押すと、/end_session に元の URL パラメータをつけて、GET する。

### /delete_cookie

cookie に有効な SESSION_ID が設定されていれば、
cookie から SESSION_ID を削除して、
URL パラメータを維持したまま /login にリダイレクト。
設定されていなければ、403 Forbidden。

(TODO) 対応する access_token 等も消すべきか。

### /delete_cookie?client_id={client_id}&redirect_uri={redirect_uri}

cookie に有効な SESSION_ID が設定されていて、
client_id が登録されているサービスの UUID で、
redirect_uri がそのサービスの提供する URI 以下ならば、
cookie から SESSION_ID を削除して、
redirect_uri にリダイレクト。
設定されていなければ、403 Forbidden。

(TODO) 対応する access_token 等も消すべきか。

----------

### /access_token?code={code}&client_id={client_id}&client_secret={client_secret}

code、client_id、client_secret が無ければ、400 Bad Request。
code が有効で、
client_id が code の発行先サービスの UUID で、
client_secret が発行先サービスが code にした署名ならば、
200 OK でボディに JSON で access_token が入る。

```
{
  "access_token": "XXXXX",
  "expires_in": 3600
}
```

そうでなければ、403 Forbidden。

----------

### /query?access_token={access_token}&client_id={client_id}&client_secret={client_secret}&attribute={attribute1}&attribute={attribute2}&...

(TODO) OpenID Connect 風にした方が良いか。

access_token、client_id、client_secret が無ければ、400 Bad Request。
access_token が有効で、
client_id が登録されているサービスの UUID で、
client_secret がそのサービスが access_token にした署名
ならば、200 OK でボディに JSON で attribute が入る。

```
{
  "user": {
    attrX: XXX,
    attrY: YYY,
    ...
  }
}
```

そうでなければ、403 Forbidden。
