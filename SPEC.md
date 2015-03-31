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

[OpenID Connect Core 1.0] も参照のこと。

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
|アクセストークン|/token|アクセストークンを発行する|
|アカウント情報|/userinfo|アカウント情報を提供する|
|TA 間連携元|/cooperation/from|TA 間連携の仲介コードを発行する|
|TA 間連携先|/cooperation/to|TA 間連携の仲介情報を提供する|
|TA 情報|/tainfo|TA の情報を提供する|


## 2. セッション

ユーザー認証、アカウント選択、ログイン、同意エンドポイントではセッションを利用する。

|Cookie 名|値|
|:--|:--|
|X-Edo-Id-Provider|セッション ID|

ユーザー認証、アカウント選択、ログイン、同意エンドポイントへのリクエスト時に、セッション ID が通知されなかった場合、セッションを発行する。
セッションの期限に余裕がない場合、設定を引き継いだセッションを発行する。

ユーザー認証、アカウント選択、ログイン、同意エンドポイントからのレスポンス時に、未通知のセッション ID を通知する。


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

エラーでない要請元 TA へのリダイレクト時には、セッションとリクエスト内容や各チケットとの紐付けを解く。
認可コード等を発行する。
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

* アカウント選択チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|アカウント選択チケット|
|**`username`**|必須|選択 / 入力されたアカウント名|
|**`locale`**|任意|選択された表示言語|

* アカウント選択チケットがセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなく、アカウント名が正当でない場合、
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
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
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

* ログインチケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|ログインチケット|
|**`username`**|必須|アカウント名|
|**`password`**|必須|入力されたパスワード|
|**`locale`**|任意|選択された表示言語|

* ログインチケットがセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなく、アカウントが正当でない場合、
    * 試行回数が閾値以下の場合、
        * ログイン UI にリダイレクトさせる。
    * そうでなければ、エラーを返す。
* そうでなければ、設定を引き継いだセッションを発行する。
  セッションをログイン済みにする。

* 同意が必要な場合（[ユーザー認証エンドポイント]を参照）、
    * 同意 UI エンドポイントにリダイレクトさせる。
* そうでなければ、要請元 TA にリダイレクトさせる。


### 5.1. リクエスト例

```http
POST /auth/login HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9
Content-Type: application/x-www-form-urlencoded

ticket=kpTK-93-AQ&username=dai.fuku&password=zYdYoFVx4sSc
```


### 5.2. レスポンス例

同意 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Set-Cookie: X-Edo-Id-Provider=GLeZi5VlD3VVxFgC-0KZQ0F0FKr0VE
    Expires=Tue, 24 Mar 2015 02:00:45 GMT; Path=/; Secure; HttpOnly
Location: /ui/consent.html?username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org#FwJrwq-8S1
```

改行とインデントは表示の都合による。


## 6. 同意エンドポイント

同意が終わった後の処理をする。

* 同意チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|同意チケット|
|**`allowed_scope`**|該当するなら必須|空白区切りの許可されたスコープ|
|**`allowed_claims`**|該当するなら必須|空白区切りの許可されたクレーム|
|**`denied_scope`**|該当するなら必須|空白区切りの拒否されたスコープ|
|**`denied_claims`**|該当するなら必須|空白区切りの拒否されたクレーム|
|**`locale`**|任意|選択された表示言語|

* 同意チケットがセッションに紐付くものと異なる、または、必要な許可が得られなかった場合、
    * エラーを返す。
* そうでなければ、要請元 TA にリダイレクトさせる。


### 6.1. リクエスト例

```http
POST /auth/consent HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=GLeZi5VlD3VVxFgC-0KZQ0F0FKr0VE
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
|**`usernames`**|任意|候補になるアカウント名の JSON 配列|
|**`display`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `display` と同じもの|
|**`locales`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `ui_locales` と同じもの|

UI の目的は、アカウント選択エンドポイントに POST させること。


### 7.1. リクエスト例

```http
GET /ui/select.html HTTP/1.1
Host: idp.example.org
```


## 8. ログイン UI エンドポイント

ログイン用の UI を提供する。

アカウント選択 UI と同じパラメータを受け付ける。

UI の目的は、ログインエンドポイントに POST させること。


### 8.1 リクエスト例

```http
GET /ui/login.html?usernames=%5B%22dai.fuku%22%5D HTTP/1.1
Host: idp.example.org
```


## 9. 同意 UI エンドポイント

同意用の UI を提供する。

以下のパラメータを受け付ける。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`username`**|必須|アカウント名|
|**`scope`**|該当するなら必須|許可が欲しいスコープ|
|**`claims`**|該当するなら必須|許可が必要なクレーム|
|**`optional_claims`**|該当するなら必須|許可が欲しいクレーム|
|**`expires_in`**|任意|発行されるアクセストークンの有効期間|
|**`client_id`**|必須|要請元 TA の ID|
|**`display`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `display` と同じもの|
|**`locales`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `ui_locales` と同じもの|

UI の目的は、同意エンドポイントに POST させること。


### 9.1. リクエスト例

```http
GET /ui/consent.html?username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org HTTP/1.1
Host: idp.example.org
```


## 10. アクセストークンエンドポイント

アクセストークンを発行する。
[OpenID Connect Core 1.0] を参照のこと。

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、認可コードと引き換えにアクセストークン等を発行する。


## 11. アカウント情報エンドポイント

アカウント情報を提供する。
[OpenID Connect Core 1.0] を参照のこと。

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなければ、アクセストークンに紐付くアカウント情報を返す。


## 12. TA 間連携元エンドポイント

TA 間連携の仲介コードを発行する。
[TA 間連携プロトコル]を参照のこと。

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、仲介コードを発行する。


## 13. TA 間連携先エンドポイント

仲介コードと引き換えに TA 間連携情報を提供する。
[TA 間連携プロトコル]を参照のこと。

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、仲介コードと引き換えに TA 間連携情報を返す。


## 14. TA 情報エンドポイント

同意 UI 用に TA の情報を返す。

TA の指定は、TA の ID をパーセントエンコードし、パスにつなげて行う。

TA 情報は以下を最上位要素として含む JSON で返される。

* **`friendly_name`**
    * 名前。
      言語タグが付くことがある。


### 14.1. リクエスト例

```http
GET /tainfo/https%3A%2F%2Fta.example.org
Host: idp.example.org
```


### 14.1. レスポンス例

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
    "friendly_name#en": "That TA",
    "friendly_name#ja": "あの TA"
}
```


## 15. エラーレスポンス

[OpenID Connect Core 1.0] と [TA 間連携プロトコル]を参照のこと。

セッションがある場合、セッションとリクエスト内容や各チケットとの紐付けを解く。


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
* 名前
* パスワード
    * ソルト
    * ハッシュ値
* 同意
* 属性
    * TA ごとに
        * 許可スコープ
        * 許可クレーム

以下の操作が必要。

* ID による取得
* 名前による取得
* TA 単位で同意の上書き

#### 16.1.2. TA 情報

以下を含む。

* ID
* 名前
* リダイレクト URI
* 検証鍵

以下の操作が必要。

* ID による取得


### 16.2. 非共有データ


#### 16.2.1. セッション

以下を含む。

* ID \*
* 有効期限 \*
* アカウント
    * ID
    * ログイン済みか \*
* リクエスト内容
* アカウント選択チケット
* ログインチケット
* 同意チケット
* 過去にログインしたアカウントの ID
* UI 表示言語

\* は設定を引き継がない。

以下の操作が必要。

* 保存
* ID による取得
* 上書き
    * ID、有効期限以外。


#### 16.2.2. 認可コード

以下を含む。

* ID
* 有効期限
* アカウント ID
* 許可スコープ
* 許可クレーム
* 要請元 TA の ID
* リダイレクト URI
* `nonce` 値
* 発行したアクセストークンの ID

以下の操作が必要。

* 保存
* ID による取得
* 発行したアクセストークンの ID の設定
    * 発行したアクセストークンの ID が未設定の場合のみ成功する。


#### 16.2.3. アクセストークン

以下を含む。

* ID
* 有効 / 無効
* 有効期限
* アカウント ID
* 許可スコープ
* 許可クレーム
* 要請元 TA の ID
* 発行したアクセストークンの ID

以下の操作が必要。

* 保存
* ID による取得
* 無効化
* 発行したアクセストークンの ID の追加
    * 有効な場合のみ成功する。


#### 16.2.4. 仲介コード

以下を含む。

* ID
* 有効期限
* 主体アカウントの ID
* 許可スコープ
* アクセストークンの有効期限
* 関連アカウントの ID
* 要請元 TA の ID
* 要請先 TA の ID
* 発行したアクセストークンの ID

以下の操作が必要。

* 保存
* ID による取得
* 発行したアクセストークンの ID の設定
    * 発行したアクセストークンの ID が未設定の場合のみ成功する。


<!-- 参照 -->
[OpenID Connect Core 1.0 Section 3.1.2.1]: http://openid-foundation-japan.github.io/openid-connect-core-1_0.ja.html#AuthRequest
[OpenID Connect Core 1.0]: http://openid.net/specs/openid-connect-core-1_0.html
[TA 間連携プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/ta_cooperation.md
[ユーザー認証エンドポイント]: #user-auth-endpoint
