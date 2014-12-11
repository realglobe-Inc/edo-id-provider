edo-id-provider
===

IdP。
アカウント認証サーバー。


起動
---

UI 用の HTML 等を ui ディレクトリの下に置く。

```
<任意のディレクトリ>/
├── edo-id-provider
└── ui
     ├── index.html
     ...
```

|オプション|値の意味・選択肢|
|:--|:--|
|-uiPath|UI 用 HTML 等を置くディレクトリパス|
|-uiUri|UI 用 HTML 等を提供する URI|


URI
---

|URI|機能|
|:--|:--|
|/login|アカウント認証する|
|/login/ui|UI 用の HTML を提供する|
|/access_token|アクセストークンを発行する|


### GET /login

prompt クエリが login、または、select_account の場合、クエリを維持したまま /ui/index.html にリダイレクトする。

そうでなく、cookie の X-Edo-Idp-Session に有効なセッションが設定されている場合、
対応するアカウントでの認証が済んでいるとみなして、POST /login の認証後処理と同じことをする。

そうでない場合、クエリを維持したまま /ui/index.html にリダイレクトする。


### POST /login

username と passwd フォームパラメータでアカウント名とパスワードを受け取り認証する。

細かい動作は OpenID Connect の OpenID Provider とほぼ同じ。
違いは認可コードの形式が一部指定されていること。


### GET /login/ui/...

UI 用の HTML を提供する。

対応するディレクトリ内に、少なくとも index.html だけは置く必要がある。

UI の役目は、最終的に、以下のように username と passwd でアカウント名とパスワードを /login に POST させること。

```
<form method="post" action="/login">
    アカウント:<input type="text" name="username" size="20" /><br/>
    パスワード:<input type="password" name="passwd" size="20" /><br/>
    <input type="submit" value="ログイン" />
</form>
```


### POST /access_token

アクセストークンを発行する。

細かい動作は OpenID Connect の OpenID Provider とほぼ同じ。
違いはトークンリクエスト時に署名によるクライアント認証を強制する点。
