edo-id-provider
===

IdP。
アカウント認証サーバー。


起動
---

UI 用の HTML 等を html ディレクトリの下に置く。

```
<任意のディレクトリ>/
├── edo-id-provider
└── html
     ├── index.html
     ...
```

|オプション|値の意味・選択肢|
|:--|:--|
|-uiPath|UI 用 HTML 等を置くディレクトリパス。初期値は実行ファイルディレクトリにある html ディレクトリ|


URI
---

|URI|機能|
|:--|:--|
|/auth/consent|同意 UI からの入力を受け付ける|
|/auth/login|ログイン UI からの入力を受け付ける|
|/auth/select|アカウント選択 UI からの入力を受け付ける|
|/auth|ユーザー認証・認可を始める|
|/html/consent.html|同意 UI 用の HTML を提供する|
|/html/login.html|ログイン UI 用の HTML を提供する|
|/html/select.html|アカウント選択 UI 用の HTML を提供する|
|/token|アクセストークンを発行する|
|/userinfo|アカウント情報を提供する|


### GET /auth

ユーザー認証・認可を始める。

動作は OpenID Provider とほぼ同じ。
違いは認可コードの形式が一部指定されていること。


### POST /auth/select

username フォームパラメータでアカウント名を受け取り、ユーザー認証・認可を続ける。


### POST /auth/login

username と passwd フォームパラメータでアカウント名とパスワードを受け取り、ユーザー認証・認可を続ける。


### POST /auth/consent

フォームパラメータで同意情報を受け取り、ユーザー認証・認可を続ける。


### GET /html/select.html#{ticket}

アカウント選択 UI 用の HTML を提供する。
目的は、以下のパラメータを /auth/select に POST させること。

|パラメータ|値|
|:--|:--|
|locale|ユーザーが選択した表示言語。必須ではない。|
|username|アカウント名|
|ticket|/auth 等からリダイレクトしたときにフラグメントで与えられる文字列|

/auth 等からリダイレクトしたときに以下のパラメータが付加される場合がある。

|パラメータ|値|
|:--|:--|
|display|画面表示形式。page/popup/touch/wap|
|locales|優先表示言語。空白区切り|
|usernames|候補となるアカウント名。JSON 配列|


### GET /html/login.html#{ticket}

ログイン UI 用の HTML を提供する。
目的は、以下のパラメータを /auth/login に POST させること。

|パラメータ|値|
|:--|:--|
|locale|ユーザーが選択した表示言語。必須ではない。|
|username|アカウント名|
|password|アカウントのパスワード|
|ticket|/auth 等からリダイレクトしたときにフラグメントで与えられる文字列|

/auth 等からリダイレクトしたときに以下のパラメータが付加される場合がある。

|パラメータ|値|
|:--|:--|
|display|画面表示形式。page/popup/touch/wap|
|locales|優先表示言語。空白区切り|
|usernames|候補となるアカウント名。JSON 配列|


### GET /html/consent.html#{ticket}

同意 UI 用の HTML を提供する。
目的は、以下のパラメータを /auth/consent に POST させること。

|パラメータ|値|
|:--|:--|
|consented_claim|同意されたクレーム。空白区切り|
|consented_scope|同意された scope。空白区切り|
|denied_claim|拒否されたクレーム。空白区切り|
|denied_scope|拒否された scope。空白区切り|
|locale|ユーザーが選択した表示言語。必須ではない。|
|ticket|/auth 等からリダイレクトしたときにフラグメントで与えられる文字列|

/auth 等からリダイレクトしたときに以下のパラメータが付加される。

|パラメータ|値|
|:--|:--|
|claim|同意が求められるクレームの空白区切りリスト|
|client_id|情報提供先 TA の ID|
|client_name|情報提供先 TA の名前|
|display|画面表示形式。page/popup/touch/wap|
|expires_in|発行されるアクセストークンの有効期間 (秒)|
|locales|優先表示言語。空白区切り|
|scope|同意が求められる scope。空白区切り|
|username|アカウント名|


### POST /token

アクセストークンを発行する。

動作は OpenID Provider とほぼ同じ。
違いはトークンリクエスト時に署名によるクライアント認証を強制する点。


#### GET /userinfo

ユーザー情報を提供する。

動作は OpenID Provider とほぼ同じ。


## ライセンス

Apache License, Version 2.0
