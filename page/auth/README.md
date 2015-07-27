<!--
Copyright 2015 realglobe, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->


# ユーザー認証機能

ユーザーを認証してアカウントを特定する。
[ユーザー認証プロトコル]も参照のこと。


## 1. 動作仕様

以降、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


### 1.1. エンドポイント

|エンドポイント名|機能|
|:--|:--|
|ユーザー認証|ユーザー認証を開始する|
|アカウント選択|アカウント選択を受け付ける|
|ログイン|ログインを受け付ける|
|同意|同意を受け付ける|
|アカウント選択 UI|アカウント選択 UI を提供する|
|ログイン UI|ログイン UI を提供する|
|同意 UI|同意 UI を提供する|


### 1.2. セッション

ユーザー認証、アカウント選択、ログイン、同意エンドポイントではセッションを利用する。

|Cookie 名|値|
|:--|:--|
|Id-Provider|セッション ID|

ユーザー認証、アカウント選択、ログイン、同意エンドポイントへのリクエスト時に、有効なセッションが宣言されなかった場合、セッションを発行する。
ユーザー認証エンドポイントでは、セッションの期限に余裕がない場合、設定を引き継いだセッションを発行する。

ユーザー認証、アカウント選択、ログイン、同意エンドポイントからのレスポンス時に、未通知のセッション ID を通知する。

||利用|新規発行|引き継ぎ発行|
|:--|:--:|:--:|:--:|
|ユーザー認証|yes|yes|yes|
|アカウント選択|yes|yes|no|
|ログイン|yes|yes|no|
|同意|yes|yse|no|
|アカウント選択 UI|-|no|no|
|ログイン UI|-|no|no|
|同意 UI|-|no|no|


### 1.3. ユーザー認証エンドポイント<a name="user-auth-endpoint" />

ユーザー認証を始める。
[OpenID Connect Core 1.0] も参照のこと。

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなければ、セッションにリクエスト内容を紐付ける。


* リクエストに `prompt` パラメータを含み、その値が `select_account` を含む場合、
    * アカウント選択 UI エンドポイントにリダイレクトさせる。
* そうでなく、リクエストに `prompt` パラメータを含み、その値が `login` を含む、または、ログイン済みのセッションでない場合、
    * ログイン UI エンドポイントにリダイレクトさせる。
* そうでなく、リクエストに `prompt` パラメータを含み、その値が `consent` を含む、または、許可が必要なスコープやクレームがある場合、
    * 同意 UI エンドポイントにリダイレクトさせる。
* そうでなければ、要請元 TA にリダイレクトさせる。

各 UI エンドポイントへのリダイレクト時には、チケットを発行する。
チケットをセッションに紐付ける。
チケットをフラグメントとして付加した UI エンドポイントにリダイレクトさせる。

エラーでない要請元 TA へのリダイレクト時には、セッションからリクエスト内容とチケットへの紐付けを解く。
認可コードや ID トークンを発行する。
それらを付加したリダイレクト URI にリダイレクトさせる。


#### 1.3.1. リクエスト例

```http
GET /auth?response_type=code%20id_token&scope=openid
    &client_id=https%3A%2F%2Fta.example.org
    &redirect_uri=https%3A%2F%2Fta.example.org%2Freturn&state=Ito-lCrO2H
    &nonce=v46QjbP6Qr&prompt=select_account HTTP/1.1
Host: idp.example.org
```

改行とインデントは表示の都合による。


#### 1.3.2. レスポンス例

アカウント選択 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Location: /ui/select.html#_GCjShrXO9
```

改行とインデントは表示の都合による。


### 1.4. アカウント選択エンドポイント

アカウントが選択された後の処理をする。

* チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|チケット|
|**`username`**|必須|選択 / 入力されたアカウントのログイン名|
|**`locale`**|任意|選択された表示言語|

* チケットがセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなく、アカウントのログイン名が正当でない場合、
    * 試行回数が閾値以下の場合、
        * アカウント選択 UI にリダイレクトさせる。
    * そうでなければ、エラーを返す。
* そうでなければ、セッションにアカウントを紐付ける。


* ログインが必要な場合（[ユーザー認証エンドポイント]を参照）、
    * ログイン UI エンドポイントにリダイレクトさせる。
* そうでなく、同意が必要な場合（[ユーザー認証エンドポイント]を参照）、
    * 同意 UI エンドポイントにリダイレクトさせる。
* そうでなければ、要請元 TA にリダイレクトさせる。


#### 1.4.1. リクエスト例

```http
POST /auth/select HTTP/1.1
Host: idp.example.org
Cookie: Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=_GCjShrXO9&username=dai.fuku
```


#### 1.4.2. レスポンス例

ログイン UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Location: /ui/login.html?usernames=%5B%22dai.fuku%22%5D#kpTK-93-AQ
```

改行とインデントは表示の都合による。


### 1.5. ログインエンドポイント

ログイン処理をする。

* チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|チケット|
|**`username`**|必須|アカウントのログイン名|
|**`passwd_type`**|必須|パスワードの形式|
|**`locale`**|任意|選択された表示言語|

パスワード形式に従って追加のパラメータも取り出す。

* チケットがセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなく、アカウントとパスワードが正当でない場合、
    * 試行回数が閾値以下の場合、
        * ログイン UI にリダイレクトさせる。
    * そうでなければ、エラーを返す。
* そうでなければ、セッションをログイン済みにする。

* 同意が必要な場合（[ユーザー認証エンドポイント]を参照）、
    * 同意 UI エンドポイントにリダイレクトさせる。
* そうでなければ、要請元 TA にリダイレクトさせる。


#### 1.5.1. パスワード形式

以下の形式を許可する。

* `STR43`


##### 1.5.1.1. STR43

以下のパラメータを追加する。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`password`**|必須|43 文字の文字列|


#### 1.5.2. リクエスト例

```http
POST /auth/login HTTP/1.1
Host: idp.example.org
Cookie: Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=kpTK-93-AQ&username=dai.fuku&passwd_type=STR43
&password=6lZEBsm-Cn9C_LzShjUHPtVAWr9xyi6akYUjnMbfDJw
```

改行は表示の都合による。


#### 1.5.3. レスポンス例

同意 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Location: /ui/consent.html?username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org#FwJrwq-8S1
```

改行とインデントは表示の都合による。


### 1.6. 同意エンドポイント

同意が終わった後の処理をする。

* チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|チケット|
|**`allowed_scope`**|該当するなら必須|空白区切りの許可されたスコープ|
|**`allowed_claims`**|該当するなら必須|空白区切りの許可されたクレーム|
|**`denied_scope`**|該当するなら必須|空白区切りの拒否されたスコープ|
|**`denied_claims`**|該当するなら必須|空白区切りの拒否されたクレーム|
|**`locale`**|任意|選択された表示言語|

* チケットがセッションに紐付くものと異なる、または、必要な許可が得られなかった場合、
    * エラーを返す。
* そうでなければ、要請元 TA にリダイレクトさせる。


#### 1.6.1. リクエスト例

```http
POST /auth/consent HTTP/1.1
Host: idp.example.org
Cookie: Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=FwJrwq-8S1&allowed_scope=openid
```


#### 1.6.2. レスポンス例

```http
HTTP/1.1 302 Found
Location: https://ta.example.org/return?code=AFnKabazoCv99dVErDtxs5RYVmwh6R
    &id_token=eyJhbGciOiJFUzI1NiJ9.eyJhdWQiOiJodHRwczovL3RhLmV4YW1wbGUub3JnIiwiY
    19oYXNoIjoibThIOGowbG5MZDZrN3FEZFNZVENqdyIsImV4cCI6MTQyNjU1ODI2MiwiaWF0IjoxN
    DI2NTU3NjYyLCJpc3MiOiJodHRwczovL2lkcC5leGFtcGxlLm9yZyIsIm5vbmNlIjoidjQ2UWpiU
    DZRciIsInN1YiI6IjE5NTA0MTYyOTc3M0FFQ0MifQ.vevlIy6dviR6Khj8XX-zJttxEbSRych8PI
    wnCQpfTttMMok2xQJu0Pgg2y5a336NOZnQLgJZgLSN4QldZb-oFA&state=Ito-lCrO2H
```

改行とインデントは表示の都合による。


### 1.7. アカウント選択 UI エンドポイント

アカウント選択用の UI を提供する。

以下のパラメータを受け付ける。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`issuer`**|任意|IdP の ID|
|**`usernames`**|任意|候補になるアカウントのログイン名の JSON 配列|
|**`display`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `display` と同じもの|
|**`locales`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `ui_locales` と同じもの|
|**`message`**|任意|ユーザー向けメッセージ|

UI の目的は、アカウント選択エンドポイントに POST させること。


#### 1.7.1. リクエスト例

```http
GET /ui/select.html HTTP/1.1
Host: idp.example.org
```


### 1.8. ログイン UI エンドポイント

ログイン用の UI を提供する。

アカウント選択 UI と同じパラメータを受け付ける。
ただし、`issuer` を必須にする。

UI の目的は、ログインエンドポイントに POST させること。
`password` の値は次のように計算する。
IdP の ID、アカウントのログイン名、入力されたパスワードをヌル文字で連結したバイト列をつくる。
それを SHA-256 ハッシュ関数の入力にする。
出力されたバイト列を Base64URL エンコードする。
できた文字列を `password` の値にする。

```
Base64URLEncode(SHA-256(<IdP の ID> || <ヌル文字> || <アカウントのログイン名> || <ヌル文字> || <パスワード>))
```


#### 1.8.1. リクエスト例

```http
GET /ui/login.html?usernames=%5B%22dai.fuku%22%5D HTTP/1.1
Host: idp.example.org
```


### 1.9. 同意 UI エンドポイント

同意用の UI を提供する。

以下のパラメータを受け付ける。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`issuer`**|任意|IdP の ID|
|**`username`**|必須|アカウントのログイン名|
|**`scope`**|該当するなら必須|許可が欲しいスコープ|
|**`claims`**|該当するなら必須|許可が必要なクレーム|
|**`optional_claims`**|該当するなら必須|許可が欲しいクレーム|
|**`expires_in`**|任意|発行されるアクセストークンの有効期間|
|**`client_id`**|必須|要請元 TA の ID|
|**`display`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `display` と同じもの|
|**`locales`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `ui_locales` と同じもの|
|**`message`**|任意|ユーザー向けメッセージ|

UI の目的は、同意エンドポイントに POST させること。


#### 1.9.1. リクエスト例

```http
GET /ui/consent.html?username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org HTTP/1.1
Host: idp.example.org
```


## 2. エラーレスポンス

[OpenID Connect Core 1.0]を参照のこと。

セッションがある場合、セッションからリクエスト内容とチケットへの紐付けを解く。


<!-- 参照 -->
[OpenID Connect Core 1.0 Section 3.1.2.1]: http://openid-foundation-japan.github.io/openid-connect-core-1_0.ja.html#AuthRequest
[OpenID Connect Core 1.0]: http://openid.net/specs/openid-connect-core-1_0.html
[ユーザー認証エンドポイント]: #user-auth-endpoint
[ユーザー認証プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/user_authentication.md
