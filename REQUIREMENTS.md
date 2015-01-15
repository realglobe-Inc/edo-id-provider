OAuth 2.0
===

+ リフレッシュトークンをサポートするなら、アクセストークンの期限が切れた後でもリフレッシュトークンによってアクセストークンを再発行できる。
+ クライアント識別子だけでクライアントを認証してはならない。
+ パブリッククライアントを認証しても良いが、信頼してはならない。
+ クライアントが一度に 2 つ以上の認証方式を利用しようとしたら拒否すべき。
+ クライアント認証用にクライアントパスワードが発行されていたら、HTTP Basic 認証をサポートしなければならない。
+ クライアント認証に client_id と client_secret を使う方法は他に手が無い場合だけにすべき。
+ クライアント認証用 client_id と client_secret がリクエスト URI に入れられていたら拒否すべき (ボディのみ許可)。
+ パスワードでクライアント認証するなら TLS 必須。
+ (パスワードによる) クライアント認証を伴うエンドポイントでは総当たり攻撃の対策をしなければならない。
+ HTTP 認証以外の認証方式を使うなら、クライアントと認証方式の紐付けを明確にしなければならない。
+ まず、リソースオーナーを認証しなければならない。
+ 認可エンドポイント URI はフラグメントを含んではならない。
+ 認可エンドポイントは HTTP GET に対応しなければならない。
+ 認可リクエストのパラメータの値が無い場合、パラメータが無いとみなさなければならない。
+ 認可リクエストの知らないパラメータは無視しなければならない。
+ 認可リクエストのパラメータが重複したら拒否すべき。
+ 認可エンドポイントのレスポンスパラメータは重複してはならない。
+ 認可リクエストに response_type パラメータが無ければ拒否しなければならない。
+ 認可リクエストの response_type パラメータの値が未知なら拒否しなければならない。
+ リダイレクトエンドポイントが絶対 URI でないなら拒否すべき。
+ リダイレクトエンドポイントにリダイレクトさせるとき、リダイレクトエンドポイントに元々付いているパラメータを維持しなければならない。
+ リダイレクトエンドポイント URI がフラグメントを含んでいたら拒否すべき。
+ リダイレクトエンドポイントが TLS 対応していない場合、リソースオーナーに警告すべき。
+ 事前登録されていないパブリッククライアントやインプリシットグラントを利用するクライアントは拒否すべき。
+ リダイレクトエンドポイント URI が複数登録されている場合、URI の一部分のみが登録されている場合、そもそも登録されていない場合、認可リクエストに redirect_uri が無ければ拒否すべき。
+ リダイレクトエンドポイント URI が無かったり、登録されているものと合致しないならリソースオーナーに知らせるべき。
+ リダイレクトエンドポイント URI が無かったり、登録されているものと合致しないならリダイレクトしてはならない。
+ トークンエンドポイント URI はフラグメントを含んではいけない。
+ トークンエンドポイントは TLS が必須。
+ HTTP POST ではないトークンリクエストは拒否すべき。
+ トークンリクエストのパラメータの値が無い場合、パラメータが無いとみなさなければならない。
+ トークンリクエストの知らないパラメータは無視しなければならない。
+ トークンリクエストのパラメータが重複したら拒否すべき。
+ トークンエンドポイントのレスポンスパラメータは重複してはならない。
+ コンフィデンシャルクライアントのクライアント認証ができないトークンリクエストは拒否すべき。
+ grant_type が authorization_code なトークンリクエストに client_id が無ければ拒否すべき。
+ アクセストークンのスコープがクライアントが要求したものと異なる場合はレスポンスに scope パラメータを含めなくてはならない。
+ 認可リクエスト時に scope パラメータが無かった場合、デフォルト値を使うか拒否しなければならない。


##### 認可コード

+ 認可リクエストに response_type と client_id が無ければ拒否する。
+ 認可レスポンスには code パラメータが必須。
+ 認可コードは 10 分以内に期限を迎えるべき。
+ 認可コードが 2 回以上使われたら拒否しなければならない。
+ 2 回以上使われた認可コードを基に発行されたアクセストークンは無効にすべき。
+ 認可リクエストに state パラメータが含まれていた場合、認可レスポンスにも state パラメータが必須。
+ 認可リクエストがリダイレクト URI が原因で失敗したら、リソースオーナーにエラーを返すべき。
+ 不正なリダイレクト URI にリダイレクトさせてはならない。
+ 認可リクエストを拒否したり、リダイレクト URI 以外が原因で失敗した場合、リダイレクト URI にクエリパラメータでエラーを返す。
+ リダイレクト URI にクエリパラメータでエラーを返す場合、error パラメータは必須。
    + 認可リクエストに必須パラメータが無い、サポート外のパラメータがある、重複パラメータがある場合、error パラメータの値は invalid_request。
    + クライアントが認可コードの発行を許可されていない場合、error パラメータの値は unauthorized_client。
    + リソースオーナーか認可サーバーがリクエストを拒否したら error パラメータの値は access_denied。
    + クライアントに依らず認可コードの発行に対応していない場合、error パラメータの値は unsupported_response_type。
    + 予期しないエラーの場合、error パラメータの値は server_error。
    + 過負荷やメンテナンスの場合、error パラメータの値は temporarily_unavailable。
+ 認可リクエストに state パラメータが含まれていた場合、リダイレクト URI にクエリパラメータでエラーを返す場合にも state パラメータが必須。
+ トークンリクエストに grant_type と code が無ければ拒否する。
+ 認可リクエストに redirect_uri が含まれており、トークンリクエストに redirect_uri が含まれていなければ拒否する。
+ クライアントが認証されていないなら、トークンリクエストに client_id が含まれていなければ拒否する。
+ トークンリクエスト時にコンフィデンシャルクライアントなら認証しなければならない。
+ 認可コードがクライアントに対して発行されたものでなければ拒否する。
+ 認可コードが正当でなければ拒否する。
+ 認可リクエストに redirect_uri が含まれていた場合、トークンリクエストの redirect_uri を一致しなければ拒否する。


##### インプリシットグラント

使わない。


##### リソースオーナーパスワードクレデンシャルグラント

使わない。


##### クライアントクレデンシャルグラント

使わない。


#### 5. アクセストークンの発行

+ トークンレスポンスは access_token と token_type を含まなければならない。
+ トークンレスポンスは expires_in を含むべき。
+ アクセストークンのスコープが認可リクエストと異なる場合、トークンレスポンスは scope を含まなければならない。
+ アクセストークンやクレデンシャルを含むレスポンスでは、Cache-Control ヘッダに no-store、Pragma ヘッダに no-cache を指定しなければならない。
+ 発行する値のサイズについて明記すべき。
+ トークンリクエストが失敗した場合、401 Unauthorized 以外は 400 Bad Request を返す。
+ トークンリクエストの失敗レスポンスには error が必須。
    + リクエストに必要なパラメータが無い、非サポートのパラメータが含まれる、パラメータが重複している、複数のクレデンシャルがある、
      複数のクライアント認証方式が使われている場合、error パラメータの値は invalid_request。
    + 未知のクライアント。クライアント認証情報が無い、認証方式が非サポート等でクライアント認証に失敗した場合、error パラメータの値は invalid_client。
    + 認可グラント (認可コード等) またはリフレッシュトークンが不正、期限切れ、失効の場合、error パラメータの値は invalid_grant。
    + リダイレクト URI が一致していない場合、error パラメータの値は invalid_grant。
    + 認可グラントが他のクライアントに発行されたものである場合、error パラメータの値は invalid_grant。
    + グラントタイプが非サポートの場合、error パラメータの値は unsupported_grant_type。
+ Authorization ヘッダを利用するクライアント認証が invalid_client で失敗した場合、401 Unauthorized を返さなければならない。
  また、その際、WWW-Authenticate ヘッダにクライアント認証を含めなければならない。


#### 6. アクセストークンの更新

+ リフレッシュトークンを発行するなら、トークンエンドポイントで更新リクエストを受け付けねばならない。
+ 更新リクエストが grant_type と refresh_token を含んでいなければ拒否すべき。
+ 更新リクエストの grant_type の値が refresh_token でなければ拒否すべき。
+ 更新リクエストの scope の値がリフレッシュトークン発行時に許可されていない値を含んでいたら拒否すべき。
+ 更新リクエストに scope がなければリフレッシュトークン発行時の scope を使わなければならない。
+ コンフィデンシャルクライアントとクライアントクレデンシャルが発行されたクライアントは認証しなければならない。
+ クライアントがリフレッシュトークンを発行したクライアントでなかったら拒否しなければならない。
+ レスポンスは発行のときと同じ。
+ 新しいリフレッシュトークンを返した場合、古いリフレッシュトークンを無効にしても良い。
+ 新しいリフレッシュトークンのスコープは更新リクエストで指定されたものと同じでなければならない。


#### 7. 保護されたリソースへのアクセス

+ リクエストを受け取ったら、アクセストークンを検証し、期限切れ、リソースがスコープの範囲外であれば拒否しなければならない。


#### 7.1. アクセストークンタイプ


#### 7.2. エラーレスポンス

+ リクエストに応えられない場合、エラーを通知すべき。
+ エラーコードを返す場合、そのパラメータ名は error にすべき。


#### 8. 仕様の拡張性


#### 8.1. アクセストークンタイプの定義

+ URI をタイプ名にする場合、ベンダー固有の実装に限定されるべき。
+ URI でないならレジストリに登録しなければならない。
+ トークンタイプが HTTP 認証方式を定義するなら、タイプ名と HTTP 認証方式名を同じにすべき。


#### 8.2. 新たなエンドポイントパラメーターの定義

+ ベンダー固有の未登録パラメータはベンダー固有の接頭辞を付けるべき。


#### 8.3. 新たな認可グラントタイプの定義


#### 8.4. 新たな認可エンドポイントレスポンスタイプの定義


#### 8.5. 追加のエラーコードの定義

+ エラーコードは勝手につくっても良い。


#### 9. ネイティブアプリケーション


#### 10. Security Considerations


#### 10.1. クライアント認証

+ ネイティブアプリケーションやユーザーエージェントベースのアプリケーションにクライアント認証目的でパスワードやクレデンシャルを発行してはならない。


#### 10.2. クライアント偽装

+ クライアント認証が可能なら、クライアント認証しなければならない。
+ リソースオーナーに認証を要求し、クライアントと認可スコープと有効期限を通知すべき。
+ リクエストが同一のクライアントから送られていることが確認できない限り、2度目以降の認可リクエストを自動的に処理すべきではない。


#### 10.3. アクセストークン

+ アクセストークンは通信経路、ストレージで機密でなければならない。
+ アクセストークンは認可サーバー、関係するリソースサーバー、発行先クライアントでのみ共有される。
+ アクセストークンの受け付けエンドポイントは TLS のサーバー認証を通さなければならない。
+ トークンの有効性を保持したままアクセストークンを生成、変更したり、トークンの生成方法を推測できないことを保証しなければならない。
+ アクセストークンのスコープを決めるときにクライアントの素性を考慮すべきである。


#### 10.4. リフレッシュトークン

+ リフレッシュトークンは通信経路、ストレージで機密でなければならない。
+ リフレッシュトークンは認可サーバーと発行先クライアントでのみ共有される。
+ リフレッシュトークンをどのクライアントに発行したか覚えておかなければならない。
+ リフレッシュトークンの受け付けエンドポイントは TLS のサーバー認証を通さなければならない。
+ クライアントを識別できるなら、リフレッシュトークンとクライアントの紐付けを検証しなければならない。
+ トークンの有効性を保持したままリフレッシュトークンを生成、変更したり、トークンの生成方法を推測できないことを保証しなければならない。


#### 10.5. 認可コード

+ 認可コードは安全な経路で送られるべき。
+ 認可コードの有効期間は短かくなければならない。
+ 認可コードは1度しか使われてはならない。
+ 認可コードが複数回使われたら、その認可コードに基き発行された全てのアクセストークンを無効にすべき。
+ クライアント認証可能なら、クライアント認証し、認可コードが発行されたクライアントかどうか確認しなければならない。


#### 10.6. 認可コードリダイレクト URI の操作

+ 認可コードの取得に用いたリダイレクト URI と認可コードとアクセストークンの交換時に表明したリダイレクト URI が一致することを確認しなければならない。
+ リダイレクト URI がリクエスト時に提供されたら、登録値と比較しなければならない。


#### 10.7. リソースオーナーパスワードクレデンシャル


#### 10.8. リクエストの機密性

+ アクセストークン、リフレッシュトークン、リソースオーナーのパスワード、クライアントのクレデンシャルは平文で送受信してはならない。
+ 認可コードは平文で送受信するべきではない。
+ state, scope パラメータの値として重要な情報を平文で含むべきではない。


#### 10.9. エンドポイントの真正性確保

+ 全てのエンドポイントで TLS サーバー認証を通さなければならない。


#### 10.10. クレデンシャルゲッシングアタック

+ アクセストークン、認可コード、リフレッシュトークン、リソースオーナーパスワード、クライアントクレデンシャルは推測されないようにしなければならない。
+ 推測できる可能性は 2^(-128) 以下にしなければならない。
+ 推測できる可能性は 2^(-160) 以下にすべき。
+ エンドユーザーのクレデンシャルは保護しなければならない。


#### 10.11. フィッシングアタック

+ エンドユーザーからのエンドポイントは全て TLS サーバー認証を通さなければならない。


#### 10.12. クロスサイトリクエストフォージェリ

+ 認可エンドポイントに CSRF 保護を実装しなければならない。
+ リソースオーナーへの通知と同意無しでクライアントが認可を取得できないことを保証しなければならない。


#### 10.13. クリックジャッキング


#### 10.14. コードインジェクションとインプットバリデーション

+ 受け取った値、得に state, redirect_uri の値をサニタイズしなければならない。
+ 受け取った値、得に state, redirect_uri の値を検証すべき。


#### 10.15. オープンリダイレクタ


#### 10.16. インプリシットフローにおけるリソースオーナーなりすましのためのアクセストークン不正利用



OpenID Connect
===


#### 1. Insroduction


#### 1.1. Requirements Notation and Conventions


#### 1.2. Terminology

+ Issuer Identifier はクエリとフラグメントは含まない。
+ Userinfo Endpoint は https で提供しなければならない。


#### 1.3. Overview


#### 2. ID Token

+ ID トークンには iss, sub, aud, exp, iat クレームが必須。
+ sub は End-User の識別子であり、発行者内で一意かつ再割り当てしてはならない。
+ aud には Relying Party の client_id を含まなければならない。
+ exp, iat の値は 1970-01-01T0:0:0Z からの秒数。
+ リクエストが max_age を含んでいたら、ID トークンに auth_time クレームは必須。
+ リクエストが nonce を含んでいたら、ID トークンに nonce クレームとしてそのまま含ませなければならない。
+ ID トークンは JWS によって署名される、または、JWS と JWE によって署名された後に暗号化されていなければならない。
+ ID トークンは直接クライアントに発行され、そのクライアントが alg=none を要求するのではない限り、alg=none を使用してはいけない。
+ ID トークンは x5u, x5c, jku, jwk ヘッダを使うべきではない。


#### 3. Authentication


#### 3.1. Authentication using the Authriation Code Flow


#### 3.1.1. Authorization Code Flow Steps


#### 3.1.2. Authorization Endpoint

+ Authorization Endpoint は TLS で提供しなければならない。


#### 3.1.2.1. Authentication Request

+ Authorization Endpoint は GET と POST をサポートしなければならない。
+ リクエストは scope, response_type, client_id, redirect_uri を含む。
+ scope の値は openid を含む。
+ 理解できない scope の値は無視しなければならない。
+ response_type の値は code。
+ redirect_uri の値が登録されているものと完全一致しない場合は拒否すべき。
+ display=page なら、認証および同意 UI を User Agent の全画面に表示すべき。
+ display=popup なら、認証および同意 UI をポップアップ表示すべき。
+ dispaly=touch なら、認証および同意 UI をタッチインターフェースに適した形で表示すべき。
+ dispaly=wap なら、認証および同意 UI を携帯電話に適した形で表示すべき。
+ prompt が none を含むなら、認証および同意 UI を表示してはならない。
+ prompt が none を含み、Client の要求に対する同意を事前に得ていないなら、login_required や interaction_required あたりのエラーを返す。
+ prompt が login を含むなら、再認証すべき。
+ prompt が login を含み、再認証できないなら、login_required あたりのエラーを返さなければならない。
+ prompt が consent を含むなら、同意を要求すべき。
+ prompt が consent を含み、同意を要求できないなら、consent_required あたりのエラーを返さなければならない。
+ prompt が select_account を含むなら、アカウント選択を要求すべき。
+ prompt が select_account を含み、アカウント選択を要求できないなら、account_selection_required あたりのエラーを返さなければならない。
+ prompt が none とその他の値を同時に含むならエラーを返さねばならない。
+ max_age があり、End-User を認証してからの経過時間が max_age の値より大きい場合、再認証しなければならない。
+ max_age がある場合、発行される ID トークンは auth_time クレームを含まなければならない。
+ ui_locales の値をサポートしていないくてもエラーにすべきではない。
+ id_token_hint の値の ID トークンに紐付く End-User が認証済み、または認証されなかった場合、login_required あたりのエラーを返す。
+ prompt=none のときに id_token_hint がなければ、invalid_request を返しても良い。
+ id_token_hint の値の ID トークンでは aud に Authrization Server が含まれている必要は無い。
+ id_token_hint の値の ID トークンは、暗号化されるなら Relying Party の鍵で暗号化される。


#### 3.1.2.2. Authentication Request Validation

+ OAuth 2.0 パラメータを OAuth 2.0 の仕様に従って検証しなければならない。
+ scope が存在し、openid を含むことを検証しなければならない。
+ 全ての必須パラメータが存在し、仕様に従って使われているかを検証しなければならない。
+ id_token_hint や claims パラメータによって提示された sub の値が認証 End-User と異なる場合は失敗しなければならない。
+ 認識できないパラメータは無視すべき。
+ エラーならそれを返さなければならない。


#### 3.1.2.3. Authorization Server Authenticates End-User

+ End-User が未認証、または、prompt が login を含む場合は認証しなければならない。
+ prompt が none を含む場合は End-User と対話してはならない。
+ prompt が none を含み、End-User が未認証で対話無しで認証できない場合、エラーを返さなければならない。
+ End-User と対話するときは、Cross-Site Request Forgery や Clickjacking に対する対策を行わなければならない。


#### 3.1.2.4. Authorization Server Obtains End-User Consent/Authorization

+ End-User を認証した後、認可決定を取得しなければならない。


#### 3.1.2.5. Successful Authentication Response

+ redirect_uri に返り値をクエリパラメータとして加えて返さなければならない。


#### 3.1.2.6. Authentication  Error Response

+ Redirection URI が無効でない限り、Redirection URI にパラメータを加えてリダイレクトで Client に返す。
+ prompt が none を含むが、対話無しでは Authentication Request が完了できない場合、error は interaction_required。
+ prompt が none を含むが、認証のための対話無しでは Authentication Request できない場合、error は login_required。
+ prompt が none を含むが、アカウント選択のための対話無しでは Authentication Request できない場合、error は account_selection_required。
+ prompt が none を含むが、同意のための対話無しでは Authentication Request できない場合、error は consent_required。
+ Authentication Request の request_uri がエラーや無効なデータを返す場合、error は invalid_request_uri。
+ Authentication Request の request が無効な Request Object の場合、error は invalid_request_object。
+ Authentication Request の request をサポートしていない場合、error は request_not_supported。
+ Authentication Request の request_uri をサポートしていない場合、error は request_uri_not_supported。
+ Authentication Request の registration をサポートしていない場合、error は registration_not_supported。
+ error は必須。
+ Authentication Request が state を含む場合、state は必須。


#### 3.1.2.7. Authentication Response Validation


#### 3.1.3. Token Endpoint

+ Token Endpoint は TLS で提供しなければならない。


#### 3.1.3.1 Token Request

+ Confidential Client を認証しなければならない。


#### 3.1.3.2. Token Request Validation

+ Client Credential を発行していたり、何らかの Client Authentication 方式を用いる場合、Client を認証しなければならない。
+ Authorization Code がその Client 向けに発行されたものか確認しなければならない。
+ Authorization Code が有効が検証しなければならない。
+ Authorization Code が以前に使用されていないことを検証すべき。
+ redirect_uri が Authorization Request の時の redirect_uri と同じことを確認しなければならない。
+ redirect_uri が無いが、登録されている redirect_uri が一つだけの場合、エラーを返さなくても良い。
+ Authorization Code が OpenID Connect Authentication Request に対して発行されたものであることを検証しなければならない。


#### 3.1.3.3. Successful Token Response

+ レスポンスに id_token を含ませなければならない。
+ トークン、シークレット、その他の機密を含むレスポンスは "Cache-Control: no-store" と "Pragma: no-cache" ヘッダフィールドを含まなければならない。


#### 3.1.3.4. Token Error Response


#### 3.1.3.5. Token Response Validation


#### 3.1.3.6. ID Token


#### 3.1.3.7. ID Token Validation

+ ID トークンを暗号化することが Relying Party との間で事前に取り決められていたなら、ID トークンを暗号化しなければならない。
+ ID トークンの iss は OpenID Provider の Issuer Identifier にしなければならない。
+ ID トークンの alg は RS256、または id_token_signed_response_alg パラメータで指定された値であるべき。
+ nonce が Authentication Request で送られたなら、nonce をそのまま ID トークンに含ませなければならない。


#### 3.1.3.8. Access Token Validation


#### 3.2. Authentication using the Implicit Flow


#### 3.3. Authentication using the Hybrid Flow


#### 4. Initiating Login from a Third Party


#### 5. Claims


#### 5.1. Standard Claims


#### 5.1.1. Address Claim


#### 5.1.2 Additional Claims

+ 独自 Claim Name は衝突耐性を持たせた方が良い。


#### 5.2. Claims Languages and Scripts

+ \#ja 等の接尾辞で言語指定した Claim Name を受け付けるべき。
+ 指定された言語の Claim Value が無いときは他で代用して良い。
+ 単一の言語を指定された Claim をその言語で返すときには Claim Name から言語指定を外すべき。


#### 5.3. UserInfo Endpoint

+ UserInfo Endpoint は TLS で提供しなければならない。
+ UserInfo Endpoint は GET と POST をサポートしなければならない。
+ UserInfo Endpoint は bearer な Access Token を受け入れなければならない。
+ UserInfo Endpoint は CORS 等で Java Script クライアントからアクセスできるようにすべき。


#### 5.3.1. UserInfo Request


#### 5.3.2. Successful UserInfo Response

+ 事前にクライアントとの間で取り決めが無いのであれば、Claim を JSON で返す。
+ 要求されたクレームの一部を返さなくても良い。
+ クレームを返さない場合は、名前自体を JSON から除くべき。
+ sub クレームを含まなければならない。
+ JSON ならレスポンスボディに入れて、Content-Type: application/json で、UTF-8 でなければならない。
+ JWT ならレスポンスボディに入れて、Content-Type: application/jwt でなければならない。
+ JWT で署名する場合は iss と aud を含むべき。


#### 5.3.3. UserInfo Error Response


#### 5.3.4. UserInfo Response Validation


#### 5.4. Requesting Claims using Scope Values

+ アクセストークンを発行したときの scope 等とは違うクレームを返しても良い。
+ scope の profile は、name, family_name, given_name, middle_name, nickname, preferred_username, profile, picture, website, gender, birthdate, zoneinfo, locale, updated_at クレームを表す。
+ scope の email は、email, email_verified クレームを表す。
+ scope の address は、address クレームを表す。
+ scope の phone は、phone_number および phone_number_verified クレームを表す。


#### 5.5. Requesting Claim using the "claims" Request Parameter

+ Authentication Request で claims パラメータを受け付けても良い。
+ claims の値は JSON。
+ Request Object に入れることもできる。
+ claims に uderinfo があるのに response_type が Access Token を発行するタイプでなかった場合、拒否すべき。
+ claims の uderinfo で要求されたクレームを返却する場合は UserInfo Endpoint から返却する。
+ claims の id_token で要求されたクレームを返却する場合は id_token で返却する。
+ claims の認識できないメンバーは無視しなければならない。


#### 5.5.1. Individual Claims Requests

+ claims で null で要求されたクレームは Voluntary Claim。
+ claims の "essential": true で要求されたクレームは Essential。
+ Essential なクレームが利用できない場合はエラーを返して良い。
+ calims の value で要求されたら、それを Claim ごとに定義されている通りに処理する。
+ calims の values で要求されたら、それを Claim ごとに定義されている通りに処理する。
+ claims の解釈不可能なメンバーは全て無視しなければならない。


#### 5.5.1.1. Requesting the "acr" Claim

+ acr が values と共に、ID トークンの Essential Claim として要求された場合、values のどれかを acr として返さなければならない。
+ Essential で要求を満たせない場合、認証に失敗したものとして扱わなければならない。
+ Essential でなく要求を満たせない場合、現在の acr を返すべき。


#### 5.5.2. Languages and Scripts for Individual Claims

+ 保持しない言語および文字種で要求されたクレームを返すときは言語タグを使用すべき。


#### 5.6. Claim Types

+ Normal Claims はサポートしなければならない。


#### 5.6.1. Normal Claims


#### 5.6.2. Aggregated and Distributed Claims


#### 5.6.2.1. Aggregated and Distributed Claims


#### 5.6.2.1. Example of Aggregated Claims


#### 5.6.2.2. Example of Distributed Claims


#### 5.7. Claim Stability and Uniqueness

+ sub と iss で End-User の安定した識別子にならなければならない。


#### 6. Passing Request Parameters as JWTs

+ Authorization Request の request パラメータを受け付けても良い。
+ Authorization Request の request_uri パラメータを受け付けても良い。


#### 6.1. Passing a Request Object by Value

+ request をサポートしていないのに Relying Party が request を送ってきたら、request_not_supported を返さなければならない。
+ request に response_type, client_id を含む場合、パラメータとして渡されたものと一致しなければならない。
+ request の中に request, request_uri を含んではならない。


#### 6.1.1. Request using the "request" Request Parameter


#### 6.2. Passing a Request Object by Reference

+ request_uri をサポートしていないのに Relying Party が request_uri を送ってきたら、request_uri_not_supported を返さなければならない。
+ request_uri に response_type, client_id を含む場合、パラメータとして渡されたものと一致しなければならない。
+ request_uri の中に request, request_uri を含んではならない。
+ フラグメントは参照先コンテンツの SHA-256 ハッシュ値の Base64URL エンコード値と考え、キャッシュの参考にすべし。

#### 6.2.1. URL Referencing the Request Object


#### 6.2.2. Request using the "request_uri" Request Parameter


#### 6.2.3. Authorization Server Fetches Request Object

+ キャッシュしていない限り、request_uri に GET リクエストを送らなければならない。


#### 6.2.4. "request_uri" Rationale


#### 6.3. Validating JWT-Based Requests


#### 6.3.1. Encrypted Request Object

+ Request Object の暗号化をさせる場合は事前に宣言しておく必要あり。
+ Request Object が暗号化されていたら復号しなければならない。
+ Request Object が署名されていたら検証すること。
+ 復号に失敗したらエラーを返さなければならない。


#### 6.3.2. Signed Request Object

+ Request Object の署名方式は登録されたものでなければならない。
+ 署名検証に失敗したらエラーを返さなければならない。


#### 6.3.3. Request Parameter Assembly and Validation

+ Request Object とそれ以外のパラメータでは Request Object の値を優先しなければならない。


#### 7. Self-Issued OpenID Provider


#### 8. Subject Identifier Types

+ 全 Client で同一の public な Subject Identifier を提供しても良い。
+ 各 Client で異なる pairwise な Subject Identifier を提供しても良い。


#### 8.1. Pairwise Identifier Algorithm

+ OpenID Provider 以外に可逆ではならない。
+ 異なる Sector Identifier は異なる Subject Indentifier にならなければならない。
+ 同じ入力に対して同じ結果でなければならない。


#### 9. Client Authentication

+ デフォルト認証方式は client_secret_basic。
+ client_secret_jwt, private_key_jwt では iss, sub, aud, jti, exp が必須。
+ 理解できない Claim は無視しなければならない。
+ JWT は client_assertion で送る。
+ client_assertion_type は urn:ietf:params:oauth:client-assertion-type:jwt-bearer


#### 10. Signatures and Encryption


#### 10.1. Signing

+ 署名者は受信者がサポートする方式を基に署名方式を選択しなければならない。
+ 公開鍵を使うなら JWK Set として公開しなければならない。
+ JWK Set に複数の鍵が含まれるなら、JWS ヘッダに kid を含まなければならない。
+ Public Client は共通鍵を使ってはならない。


#### 10.2. Encryption


#### 11. Offline Access

+ scope が offline_access を含むとき、prompt に consent が必須。
+ Refresh Token を発行するには明示的同意が必須。事前同意では十分でないことも。
+ scope が offline_access を含み、prompt が consent を含まない、または、
  何か他の offline access 要求処理の条件を満たしていない場合、offline_access を無視しなければならない。
+ Authorization Code を返すような response_type が指定されていない場合、offline_access を無視しなければならない。
+ Client の application_type が web ならば、明示的同意が必要。
+ Client の application_type が native ならば、明示的に同意させるべき。


#### 12. Using Refresh Token


#### 12.1. Refresh Request

+ Client を認証しなければならない。
+ Refresh Token が Client に対して発行されたものであることを確認しなければならない。


#### 12.2. Successful Refresh Response

+ ID Token も再発行する場合、iss は最初の iss と同じでなければならない。
+ ID Token も再発行する場合、sub は最初の sub と同じでなければならない。
+ ID Token も再発行する場合、iat は更新しなければならない。
+ ID Token も再発行する場合、aud は最初の aud と同じでなければならない。
+ ID Token も再発行し auth_time を含める場合、最初の認証を行った時刻でなければならない。


#### 12.3. Refresh Error Response


#### 13. Serializations


#### 13.1. Query String Serialization


#### 13.2. Form Serialization


#### 13.3. JSON Serialization

+ 省略されているパラメータ、値が指定されていないパラメータは省き、null にしないべき。


#### 14. String Operations

+ Unicode Normalization してはならない。
+ スペース区切りで文字列リストを表現する場合は、単一のスペース文字 (0x20) を使用しなければならない。


#### 15. Implementation Considerations


#### 15.1. 全ての OpenID プロバイダが実装必須な機能について

+ ID Tokens を RS256 で署名できなければならない。
+ prompt による挙動変更をサポートしなければならない。
+ display パラメータをサポートしなければならない。最低限エラーにならなければ良い。
+ ui_locales と claims_locales パラメータをサポートしなければならない。最低限エラーにならなければ良い。
+ auth_time クレームを返却できるようにしなければならない。
+ max_age パラメータで認証切れを強制できるようにしなければならない。
+ acr_values パラメータをサポートしなければならない。最低限エラーにならなければ良い。


#### 15.2. Mandatory to Implement Features for Dynamic OpenID Providers

+ 事前設定していない Relying Party でも動作するなら、レスポンスタイプ id_token, code, id_token token をサポートしなければならない。
+ 事前設定していない Relying Party でも動作するなら、Discovery 機能をサポートしなければならない。
+ 事前設定していない Relying Party でも動作するなら、Dynamic Client Registration をサポートしなければならない。
+ 事前設定していない Relying Party でも動作するなら、UserInfo エンドポイントをサポートしなければならない。
+ 事前設定していない Relying Party でも動作するなら、公開鍵を素の JWK として公開しなければならない。
+ 事前設定していない Relying Party でも動作するなら、request_uri をサポートしなければならない。


#### 15.3. Discovery and Registration


#### 15.4. Mandatory to Implement Features for Relying Parties


#### 15.5. Implementation Notes


#### 15.6. Compatibility Notes


#### 15.7. Related Specifications and Implementer's Guides


#### 16. Security Considerations


#### 16.1. Request Disclosure


#### 16.2. Server Masquerading


#### 16.3. Token Manifacture/Modification


#### 16.4. Access Token Disclosure

+ Access Token は認可されていない対象にさらされることはあってはならない。


#### 16.5. Server Response Disclosure


#### 16.6. Server Response Repudiation


#### 16.7. Request Repudiation

+ Clinet に署名されたリクエストを要求しても良い。
+ 署名は検証すべき。


#### 16.8. Access Token Redirect

+ Access Token は audience と scope に対して制限をかけるべき。
+ audience にリソースの識別子を加えても良い。


#### 16.9. Token Reuse

+ トークンはタイムスタンプを含み、有効期限を短くすべき。


#### 16.10. Eavesdropping or Leaking Authorization Codes (Secondary Authenticator Capture)


#### 16.11. Token Substitution

+ ID Token の c_hash, at_hash で Authorization Code 置換、Access Token 置換を防げる。


#### 16.12. Timing Attack

+ エラーを検知してもすぐに検証を終わらせず、他のオクテットに対する処理を終えるまで継続すべき。


#### 16.13. Other Crypto Related Attacks


#### 16.14. Signing and Encryption Order

+ 暗号化された JWT に署名する必要は無い。


#### 16.15. Issuer Identifier

+ 1 つの Host と Port のペアに複数の Issuer がいて良い。
+ Host ごとに単一の Issuer を推奨。


#### 16.16. Implicit Flow Threats


#### 16.17. TLS Requirements

+ TLS をサポートしなければならない。
+ TLS を利用する場合、サーバー証明書を検証しなければならない。

#### 16.18. Lifetimes of Access Tokens and Refresh Tokens

+ Access Token の有効期間は非常に短い期間に限定すべき。
+ Refresh Token などにより長期間の認可を与える場合、End-User に明示すべき。
+ End-User に対し、Access token や Refresh Token を無効化する手段を提供すべき。


#### 16.19. Symmetric Key Entropy

+ client_secret は十分に強固な鍵でなければならない。


#### 16.20. Need for Signed Requests


#### 16.21. Need for Encrypted Requests


#### 17. Privacy Considerations


#### 17.1. Personally Identifiable Information


#### 17.2. Data Access Monitoring

+ UserInfo のアクセスログを残すべき。


#### 17.3. Correlation

+ 名寄せを避けるため、Pairwise Pseudonymous Identifier を使うべき。


#### 17.4. Offline Access

+ Offline Access のときは同意をしっかりさせるべき。
