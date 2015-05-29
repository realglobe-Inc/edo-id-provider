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


# edo-id-provider の仕様（目標）

[ユーザー認証プロトコル]と [TA 間連携プロトコル]も参照のこと。

以降の動作記述において、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


## 1. エンドポイント

|エンドポイント名|初期パス|機能|
|:--|:--|:--|
|ユーザー認証|/auth|ユーザー認証を開始する|
|アカウント選択|/auth/select|アカウント選択を受け付ける|
|ログイン|/auth/login|ログインを受け付ける|
|同意|/auth/consent|同意を受け付ける|
|アカウント選択 UI|/ui/select.html|アカウント選択 UI を提供する|
|ログイン UI|/ui/login.html|ログイン UI を提供する|
|同意 UI|/ui/consent.html|同意 UI を提供する|
|TA 情報|/api/info/ta|UI 用に TA の情報を提供する|
|アクセストークン|/api/token|アクセストークンを発行する|
|アカウント情報|/api/info/account|アカウント情報を提供する|
|TA 間連携元|/api/coop/from|TA 間連携の仲介コードを発行する|
|TA 間連携先|/api/coop/to|TA 間連携の仲介情報を提供する|


## 2. セッション

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
|TA 情報|-|no|no|
|アクセストークン|no|no|no|
|アカウント情報|no|no|no|
|TA 間連携元|no|no|no|
|TA 間連携先|no|no|no|


## 3. ユーザー認証エンドポイント<a name="user-auth-endpoint" />

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


### 3.1. リクエスト例

```http
GET /auth?response_type=code%20id_token&scope=openid
    &client_id=https%3A%2F%2Fta.example.org
    &redirect_uri=https%3A%2F%2Fta.example.org%2Freturn&state=Ito-lCrO2H
    &nonce=v46QjbP6Qr&prompt=select_account HTTP/1.1
Host: idp.example.org
```

改行とインデントは表示の都合による。


### 3.2. レスポンス例

アカウント選択 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Location: /ui/select.html#_GCjShrXO9
```

改行とインデントは表示の都合による。


## 4. アカウント選択エンドポイント

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


### 4.1. リクエスト例

```http
POST /auth/select HTTP/1.1
Host: idp.example.org
Cookie: Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=_GCjShrXO9&username=dai.fuku
```


### 4.2. レスポンス例

ログイン UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Location: /ui/login.html?usernames=%5B%22dai.fuku%22%5D#kpTK-93-AQ
```

改行とインデントは表示の都合による。


## 5. ログインエンドポイント

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


### 5.1. パスワード形式

以下の形式を許可する。

* `STR43`


#### 5.1.1. STR43

以下のパラメータを追加する。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`password`**|必須|43 文字の文字列|


### 5.2. リクエスト例

```http
POST /auth/login HTTP/1.1
Host: idp.example.org
Cookie: Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=kpTK-93-AQ&username=dai.fuku&passwd_type=STR43
&password=6lZEBsm-Cn9C_LzShjUHPtVAWr9xyi6akYUjnMbfDJw
```

改行は表示の都合による。


### 5.3. レスポンス例

同意 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Location: /ui/consent.html?username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org#FwJrwq-8S1
```

改行とインデントは表示の都合による。


## 6. 同意エンドポイント

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


### 6.1. リクエスト例

```http
POST /auth/consent HTTP/1.1
Host: idp.example.org
Cookie: Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=FwJrwq-8S1&allowed_scope=openid
```


### 6.2. レスポンス例

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


## 7. アカウント選択 UI エンドポイント

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


### 7.1. リクエスト例

```http
GET /ui/select.html HTTP/1.1
Host: idp.example.org
```


## 8. ログイン UI エンドポイント

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


### 8.1. リクエスト例

```http
GET /ui/login.html?usernames=%5B%22dai.fuku%22%5D HTTP/1.1
Host: idp.example.org
```


## 9. 同意 UI エンドポイント

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


### 9.1. リクエスト例

```http
GET /ui/consent.html?username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org HTTP/1.1
Host: idp.example.org
```


## 10. TA 情報エンドポイント

同意 UI 用に TA の情報を返す。

TA の指定は、TA の ID をパーセントエンコードし、パスにつなげて行う。

TA 情報は以下を最上位要素として含む JSON で返される。

* **`client_name`**
    * 名前。
      言語タグが付くことがある。


### 10.1. リクエスト例

```http
GET /api/info/ta/https%3A%2F%2Fta.example.org
Host: idp.example.org
```


### 10.2. レスポンス例

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
    "client_name#en": "That TA",
    "client_name#ja": "あの TA"
}
```


## 11. アクセストークンエンドポイント

アクセストークンを発行する。
[OpenID Connect Core 1.0] を参照のこと。

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、認可コードと引き換えにアクセストークン等を発行する。


## 12. アカウント情報エンドポイント

アカウント情報を提供する。
[OpenID Connect Core 1.0] を参照のこと。

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなければ、アクセストークンに紐付くアカウント情報を返す。


## 13. TA 間連携元エンドポイント

TA 間連携の仲介コードを発行する。
[TA 間連携プロトコル]を参照のこと。

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、仲介コードを発行する。


## 14. TA 間連携先エンドポイント

仲介コードと引き換えに TA 間連携情報を提供する。
[TA 間連携プロトコル]を参照のこと。

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、仲介コードと引き換えに TA 間連携情報を返す。


## 15. エラーレスポンス

[OpenID Connect Core 1.0] と [TA 間連携プロトコル]を参照のこと。

セッションがある場合、セッションからリクエスト内容とチケットへの紐付けを解く。


## 16. 外部データ

以下に分ける。

* 共有データ
    * 他のプログラムと共有する可能性のあるもの。
* 非共有データ
    * 共有するとしてもこのプログラムの別プロセスのみのもの。


### 16.1. 共有データ


#### 16.1.1. アカウント情報

以下を含む。

* ID
* ログイン名
* パスワード
    * パスワード形式
    * ソルト（`STR43` なら）
    * ハッシュ値（`STR43` なら）
* 属性

以下の操作が必要。

* ID による取得
* ログイン名による取得


#### 16.1.2. 同意情報

管理 UI と共有する。
以下を含む。

* アカウント ID
* TA 集
    * ID
    * 許可スコープ
    * 許可属性（クレーム）

以下の操作が必要。

* TA 単位で同意の上書き


#### 16.1.3. TA 情報

以下を含む。

* ID
* 表示名
* リダイレクト URI
* 検証鍵

以下の操作が必要。

* ID による取得


#### 16.1.4. ID プロバイダ情報

以下を含む。

* ID
* 検証鍵

以下の操作が必要。

* ID による取得


### 16.2. 非共有データ


#### 16.2.1. TA 固有のアカウント ID

TA 固有のアカウント ID を使う場合のみ。
TA 間連携プロトコルで逆引きが必要になる。

以下を含む。

* アカウント ID
* TA の ID
* TA 固有のアカウント ID

以下の操作が必要。

* 保存
* TA の ID と TA 固有のアカウント ID による取得
* TA の ID による削除
* アカウント ID による削除


#### 16.2.2. セッション

以下を含む。

* ID \*
* 有効期限 \*
* アカウント
    * ID
    * ログイン名
    * ログイン済みか \*
* リクエスト内容 \*
* チケット \*
* 過去にログインしたアカウント集
    * ID
    * ログイン名
    * ログイン済みか \*
* UI 表示言語

\* は設定を引き継がない。

以下の操作が必要。

* 保存
* ID による取得
* 上書き
    * ID、有効期限以外。


#### 16.2.3. 認可コード

以下を含む。

* ID
* 有効期限
* アカウント ID
* 許可スコープ
* 許可属性
* 要請元 TA の ID
* リダイレクト URI
* `nonce` 値
* 発行したアクセストークンの ID

以下の操作が必要。

* 保存
* ID による取得
* 発行したアクセストークンの ID の設定
    * 発行したアクセストークンの ID が未設定の場合のみ成功する。


#### 16.2.4. アクセストークン

以下を含む。

* ID
* 有効 / 無効
* 有効期限
* アカウント ID
* 許可スコープ
* 許可属性
* 要請元 TA の ID
* 発行したアクセストークンの ID

以下の操作が必要。

* 保存
* ID による取得
* 無効化
* 発行したアクセストークンの ID の追加
    * 有効な場合のみ成功する。


#### 16.2.5. 仲介コード

以下を含む。

* ID
* 有効期限
* 主体アカウント
    * ID
    * タグ
* 許可スコープ
* アクセストークンの有効期限
* 関連アカウント集
    * ID
    * タグ
* 要請元 TA の ID
* 要請先 TA の ID
* 発行したアクセストークンの ID

以下の操作が必要。

* 保存
* ID による取得
* 発行したアクセストークンの ID の設定
    * 発行したアクセストークンの ID が未設定の場合のみ成功する。


#### 16.2.6. JWT の ID

JWT の再利用を防ぐため。

以下を含む。

* `iss`
* `jti`
* `exp` または有効期限

以下の操作が必要。

* 保存
    * `iss` と `jti` が重複していたら失敗する。


<!-- 参照 -->
[OpenID Connect Core 1.0 Section 3.1.2.1]: http://openid-foundation-japan.github.io/openid-connect-core-1_0.ja.html#AuthRequest
[OpenID Connect Core 1.0]: http://openid.net/specs/openid-connect-core-1_0.html
[TA 間連携プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/ta_cooperation.md
[ユーザー認証エンドポイント]: #user-auth-endpoint
[ユーザー認証プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/user_authentication.md
