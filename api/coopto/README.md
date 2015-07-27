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


# 連携先 TA 間連携仲介機能

仲介コードと引き換えに TA 間連携情報を提供する。
[TA 間連携プロトコル]を参照のこと。


## 1. 動作仕様

* リクエストに問題がある、または、要請元 TA に問題がある場合、
    * エラーを返す。
* そうでなければ、仲介コードと引き換えに TA 間連携情報を返す。


<!-- 参照 -->
[TA 間連携プロトコル]: https://github.com/realglobe-Inc/edo/blob/master/ta_cooperation.md
