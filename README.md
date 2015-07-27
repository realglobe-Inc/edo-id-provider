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


# edo-id-provider

[EDO] の ID プロバイダ。


## 1. インストール

[go] が必要。
go のインストールは http://golang.org/doc/install を参照のこと。

go をインストールしたら、

```shell
go get github.com/realglobe-Inc/edo-id-provider
```

適宜、依存ライブラリを `go get` すること。


## 2. 実行

以下ではバイナリファイルが `${GOPATH}/bin/edo-id-provider` にあるとする。
パスが異なる場合は置き換えること。


### 2.1. DB の準備

キャッシュやセッション等に [redis]、ID プロバイダ・TA・アカウント情報等に [mongodb] が必要になる。

mongodb への ID プロバイダ・TA・アカウント情報等の同期は別口で行う。


### 2.2. UI の準備

UI を edo-id-provider で提供する場合は、適当なディレクトリに UI 用ファイルを用意する。

```
<UI ディレクトリ>/
├── consent.html
├── login.html
├── select.html
...
```

UI ディレクトリは起動オプションで指定する。


### 2.3. 起動

単独で実行できる。

```shell
${GOPATH}/bin/edo-idp-selector
```

### 2.4. 起動オプション

|オプション名|初期値|値|
|:--|:--|:--|
|-uiDir||UI 用ファイルを置くディレクトリパス|


### 2.5. デーモン化

単独ではデーモンとして実行できないため、[Supervisor] 等と組み合わせて行う。


## 3. 動作仕様

ユーザー認証および TA 間連携の仲介を行う。


### 3.1. エンドポイント

|エンドポイント名|初期パス|機能|
|:--|:--|:--|
|ユーザー認証|/auth|[ユーザー認証機能](/page/auth)を参照|
|アカウント選択|/auth/select|[ユーザー認証機能](/page/auth)を参照|
|ログイン|/auth/login|[ユーザー認証機能](/page/auth)を参照|
|同意|/auth/consent|[ユーザー認証機能](/page/auth)を参照|
|アカウント選択 UI|/ui/select.html|[ユーザー認証機能](/page/auth)を参照|
|ログイン UI|/ui/login.html|[ユーザー認証機能](/page/auth)を参照|
|同意 UI|/ui/consent.html|[ユーザー認証機能](/page/auth)を参照|
|TA 情報|/api/info/ta|[TA 情報提供機能](https://github.com/realglobe-Inc/edo-idp-selector/blob/master/api/ta)を参照|
|アクセストークン|/api/token|[アクセストークン発行機能](/api/token)を参照|
|アカウント情報|/api/info/account|[アカウント情報提供機能](/api/account)を参照|
|TA 間連携元|/api/coop/from|[連携元用 TA 間連携仲介機能](/api/coopfrom)を参照|
|TA 間連携先|/api/coop/to|[連携先用 TA 間連携仲介機能](/api/coopto)を参照|


## 4. API

[GoDoc](http://godoc.org/github.com/realglobe-Inc/edo-idp-selector)


## 5. ライセンス

Apache License, Version 2.0


<!-- 参照 -->
[EDO]: https://github.com/realglobe-Inc/edo/
[Supervisor]: http://supervisord.org/
[go]: http://golang.org/
[mongodb]: https://www.mongodb.org/
[redis]: http://redis.io/
