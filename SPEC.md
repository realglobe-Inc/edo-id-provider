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


## 1. エンドポイント

|エンドポイント名|初期 URI|機能|
|:--|:--|:--|
|ユーザー認証|/auth|ユーザー認証を開始する|
|アカウント選択|/auth/select|アカウント選択を受け付ける|
|ログイン|/auth/login|ログインを受け付ける|
|同意|/auth/consent|同意を受け付ける|
|アカウント選択 UI|/html/select.html|アカウント選択 UI を提供する|
|ログイン UI|/html/login.html|ログイン UI を提供する|
|同意 UI|/html/consent.html|同意 UI を提供する|
|アクセストークン|/token|アクセストークンを発行する|
|ユーザー情報|/userinfo|ユーザー情報を提供する|
|TA 間連携元|/cooperation/from|TA 間連携の仲介コードを発行する|
|TA 間連携先|/cooperation/to|TA 間連携情報を提供する|


### 1.1. 記法

以降の動作仕様の記述において、箇条書きに以下の構造を持たせることがある。

* if
    * then
* else if
    * then
* else


## 2. ユーザー認証エンドポイント

ユーザー認証を始める。
[OpenID Connect Core 1.0] も参照のこと。

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなく、`prompt` パラメータにて強制される処理がある場合、
    * 対応する UI エンドポイントにリダイレクトさせる。
* そうでなく、Cookie に X-Edo-Id-Provider を含み、その値が有効なセッションの場合、
    * 認可コードを発行し、要請元 TA にリダイレクトさせる。
* そうでなければ、必要な UI エンドポイントにリダイレクトさせる。

|Cookie ラベル|値|
|:--|:--|
|X-Edo-Id-Provider|セッション ID|

UI エンドポイントへのリダイレクト時には、チケットを発行し、それをセッションに紐付け、セッションを更新しつつ、チケットをフラグメントとして付加した UI エンドポイントにリダイレクトさせる。

要請元 TA へのリダイレクト時には、セッションを更新しつつ、リダイレクトさせる。


### 2.1. リクエスト例

```http
GET /auth?response_type=code%20id_token&scope=openid
    &client_id=https%3A%2F%2Fta.example.org
    &redirect_uri=https%3A%2F%2Fta.example.org%2Freturn
    &state=Ito-lCrO2H&nonce=v46QjbP6Qr HTTP/1.1
Host: idp.example.org
```


### 2.2. レスポンス例

アカウント選択 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Set-Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l;
    Path=/; Expires=Tue, 24 Mar 2015 01:59:25 GMT
Location: /html/select.html#_GCjShrXO9
```

改行とインデントは表示の都合による。


## 3. アカウント選択エンドポイント

アカウントが選択された後の処理をする。

* Cookie に X-Edo-Id-Provider を含まない、または、アカウント選択チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|アカウント選択チケット|
|**`username`**|必須|選択されたアカウント名|

* アカウント選択チケットがセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなく、アカウント名が正当でない場合、
    * 試行回数が閾値以下の場合、
        * アカウント選択 UI にリダイレクトさせる。
    * そうでなければ、エラーを返す。
* そうでなければ、セッションにアカウントを紐付ける。


* ログインまたは同意が必要な場合、
    * 対応する UI エンドポイントにリダイレクトさせる。
* そうでなければ、認可コードを発行し、要請元 TA にリダイレクトさせる。


### 3.1. リクエスト例

```http
POST /auth/select HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l
Content-Type: application/x-www-form-urlencoded

ticket=_GCjShrXO9&username=dai.fuku
```


### 3.2. レスポンス例

ログイン UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Host: idp.example.org
Set-Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l;
    Path=/; Expires=Tue, 24 Mar 2015 02:00:08 GMT
Location: /html/login.html?usernames=%5B%22dai.fuku%22%5D#kpTK-93-AQ
```

改行とインデントは表示の都合による。


## 4. ログインエンドポイント

ログイン処理をする。

* Cookie に X-Edo-Id-Provider を含まない、または、ログインチケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|ログインチケット|
|**`username`**|必須|アカウント名|
|**`password`**|必須|入力されたパスワード|

* ログインチケットがセッションに紐付くものと異なる場合、
    * エラーを返す。
* そうでなく、アカウントが正当でない場合、
    * 試行回数が閾値以下の場合、
        * ログイン UI にリダイレクトさせる。
    * そうでなければ、エラーを返す。
* そうでなければ、セッションにログイン済みアカウントを紐付ける。

* 同意が必要な場合、
    * 同意 UI エンドポイントにリダイレクトさせる。
* そうでなければ、認可コードを発行し、要請元 TA にリダイレクトさせる。


### 4.1. リクエスト例

```http
POST /auth/login HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l
Content-Type: application/x-www-form-urlencoded

ticket=kpTK-93-AQ&username=dai.fuku&password=zYdYoFVx4sSc
```


### 4.2. レスポンス例

同意 UI へのリダイレクト例。

```http
HTTP/1.1 302 Found
Host: idp.example.org
Set-Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l;
    Path=/; Expires=Tue, 24 Mar 2015 02:00:45 GMT
Location: /html/consent.html?
    username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org
    &client_friendly_name=%E4%BD%95%E3%81%8B%E3%81%AE%20TA
    #FwJrwq-8S1
```

改行とインデントは表示の都合による。


## 5. 同意エンドポイント

同意が終わった後の処理をする。

* Cookie に X-Edo-Id-Provider を含まない、または、同意チケットと紐付くセッションでない場合、
    * エラーを返す。
* そうでなければ、リクエストから以下のパラメータを取り出す。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`ticket`**|必須|同意チケット|
|**`consented_scope`**|該当するなら必須|提供の同意が得られた空白区切りのスコープ|
|**`consented_claims`**|該当するなら必須|提供の同意が得られた空白区切りのクレーム|
|**`denied_scope`**|該当するなら必須|提供が拒否された空白区切りのスコープ|
|**`denied_claims`**|該当するなら必須|提供が拒否された空白区切りのクレーム|

* 同意チケットがセッションに紐付くものと異なる、または、必要な同意が得られなかった場合、
    * エラーを返す。
* そうでなければ、認可コードを発行し、セッションを更新しつつ、要請元 TA にリダイレクトさせる。


### 5.1. リクエスト例

```http
POST /auth/consent HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l
Content-Type: application/x-www-form-urlencoded

ticket=FwJrwq-8S1&consented_scope=openid
```


### 5.2. レスポンス例

```http
HTTP/1.1 302 Found
Set-Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l;
    Path=/; Expires=Tue, 24 Mar 2015 02:01:02 GMT
Location: https://ta.example.org/return?
    code=AFnKabazoCv99dVErDtxs5RYVmwh6R6VoP7Fygrw
    &id_token=...
    &state=Ito-lCrO2H
```


## 6. アカウント選択 UI エンドポイント

アカウント選択用の UI を提供する。

以下のパラメータを受け付ける。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`usernames`**|任意|候補になるアカウント名の JSON 配列|
|**`display`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `display` と同じもの|
|**`locales`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `ui_locales` と同じもの|

UI の目的は、アカウント選択エンドポイントに POST させること。


### 6.1. リクエスト例

```http
POST /html/select.html HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l
```


## 7. ログイン UI エンドポイント

ログイン用の UI を提供する。

アカウント選択 UI と同じパラメータを受け付ける。

UI の目的は、ログインエンドポイントに POST させること。


### 7.1 リクエスト例

```http
POST /html/login.html?usernames=%5B%22dai.fuku%22%5D HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l
```


## 8. 同意 UI エンドポイント

同意用の UI を提供する。

以下のパラメータを受け付ける。

|パラメータ名|必要性|値|
|:--|:--|:--|
|**`username`**|必須|アカウント名|
|**`scope`**|該当するなら必須|同意が求められるスコープ|
|**`claims`**|該当するなら必須|同意が求められるクレーム|
|**`expires_in`**|任意|発行されるアクセストークンの有効期間|
|**`client_id`**|必須|要請元 TA の ID|
|**`client_friendly_name`**|必須|要請元 TA の名前|
|**`display`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `display` と同じもの|
|**`locales`**|任意|[OpenID Connect Core 1.0 Section 3.1.2.1] の `ui_locales` と同じもの|

UI の目的は、同意エンドポイントに POST させること。


### 8.1. リクエスト例

```http
POST /html/consent.html?
    username=dai.fuku&scope=openid&expires_in=3600
    &client_id=https%3A%2F%2Fta.example.org
    &client_friendly_name=%E4%BD%95%E3%81%8B%E3%81%AE%20TA HTTP/1.1
Host: idp.example.org
Cookie: X-Edo-Id-Provider=gxQyExhR8QojI0Cxx-JVWIhhf_5Ac9k1AHbQO62l
```


## 9. アクセストークンエンドポイント

アクセストークンを発行する。
[OpenID Connect Core 1.0] を参照のこと。


## 10. ユーザー情報エンドポイント

ユーザー情報を提供する。
[OpenID Connect Core 1.0] を参照のこと。


## 11. TA 間連携元エンドポイント

TA 間連携の仲介コードを発行する。
[TA 間連携プロトコル]を参照のこと。


## 12. TA 間連携先エンドポイント

仲介コードと引き換えに TA 間連携情報を提供する。
[TA 間連携プロトコル]を参照のこと。


## エラーレスポンス

[OpenID Connect Core 1.0] と [TA 間連携プロトコル]を参照のこと。


## 外部データ

以下に分ける。

* 共有データ
    * 他のプログラムと共有する可能性のあるもの。
* 非共有データ
    * 共有するとしてもこのプログラムの別プロセスのみのもの。


### 共有データ

* アカウント
    * ID
    * 名前
    * パスワード情報（ソルトとハッシュ値）
    * 同意
* TA
    * ID
    * 名前
    * リダイレクト URI
    * 検証鍵


### 非共有データ

* セッション
    * ID
    * 有効 / 無効
    * 過去にログインしたアカウントの ID
    * 現在の認証リクエスト
    * アカウント選択チケット
    * ログインチケット
    * 同意チケット
* 認可コード
    * ID
    * 有効期限
    * スコープ
    * クレーム
    * ユーザー ID
    * 提供先 TA の ID


<!-- 参照 -->
[OpenID Connect Core 1.0 Section 3.1.2.1]: http://openid-foundation-japan.github.io/openid-connect-core-1_0.ja.html#AuthRequest
[OpenID Connect Core 1.0]: http://openid.net/specs/openid-connect-core-1_0.html
[TA 間連携プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/ta_cooperation.md
