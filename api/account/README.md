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


# アカウント情報提供機能

アカウント情報を提供する。
[OpenID Connect Core 1.0] を参照のこと。


## 1. 動作仕様

* リクエストに問題がある場合、
    * エラーを返す。
* そうでなければ、アクセストークンに紐付くアカウント情報を返す。


<!-- 参照 -->
[OpenID Connect Core 1.0]: http://openid.net/specs/openid-connect-core-1_0.html
