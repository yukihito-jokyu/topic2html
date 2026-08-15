# FEAT-001 実装タスク — 管理者本人確認と管理アクセス

## Task一覧

| ID | タイトル | 論理領域 | 依存 |
| --- | --- | --- | --- |
| TASK-001-01 | Backend／Frontend分離と4層認証基盤の構造移行を行う | 実行基盤・構造境界 | なし |
| TASK-001-02 | 認証設定とGoogle外部境界を実装する | 設定・外部OAuth境界 | TASK-001-01 |
| TASK-001-03 | 管理認証の保護記録とmigrationを実装する | 永続化・migration | TASK-001-01 |
| TASK-001-04 | Google OAuth/OIDCによる管理session発行を実装する | Application・認証・外部連携 | TASK-001-02、TASK-001-03 |
| TASK-001-05 | 管理session・CSRF・共通アクセスガードを実装する | Application・認可・セキュリティ | TASK-001-02、TASK-001-03 |
| TASK-001-06 | 管理認証画面の状態と操作を実装する | presentation | TASK-001-01、TASK-001-04、TASK-001-05 |
| TASK-001-07 | 認証契約と構造境界の検証を自動化する | testing / verification | TASK-001-01〜06 |

## TASK-001-01 Backend／Frontend分離と4層認証基盤の構造移行を行う

**目的:** 信頼済み認証アプリケーションをBackend、管理UIをFrontendとして分離し、Backendを`handler`、`usecase`、`repository`、`domain`および`cmd`へ明示配置する。既存の認証設定・Google境界を安全に維持したまま、以降の認証機能を承認済みの依存規則に従って実装できる基盤を成立させる。

**関連要件:** REQ-026、REQ-027、BR-015、NFR-001、NFR-003

**作業内容:** Backendを独立したGo依存・検証単位へ移行し、Frontendを独立した管理UIの作業領域として成立させる。Backendでは`cmd`をcomposition root、`handler`をGin HTTP境界、`usecase`を認証操作とport、`repository`をPostgreSQL・Google・暗号などの外部I/O、`domain`を業務規則として分離する。既存の安全な起動時設定検証は`cmd`に、Google外部通信は`repository`に置き、後続Taskで認証機能を拡張できる境界を成立させる。

**受け入れ基準:**

- `backend/`と`frontend/`の責務が[Clean Architecture境界](design/architecture-boundaries.md)に一致し、`backend/`には`cmd`、`handler`、`usecase`、`repository`、`domain`が明示配置される。Frontendは同一originのHTTP契約以外のBackend内部、PostgreSQL、Google、秘密情報へ依存しない。
- Backendは独立したGo moduleおよび検証単位となり、Go依存はBackendだけで管理される。旧来のroot Go moduleがBackend依存の正本として残らない。
- `cmd`は設定読取り・起動時検証・依存の組み立てに限定される。`domain`は外部層へ、`usecase`はGin、PostgreSQL、Google、環境設定、`repository`実装へ、`handler`は`repository`へ、`repository`は`handler`へ直接依存しない。
- Ginは`handler`でHTTP route、middleware、request/response、redirect、cookieの契約変換だけを担当し、Domain規則、認可判定、Google claim検証、SQL実行を重複して保持しない。
- 構造移行中も、設定不備では認証機能を提供せず安全に起動失敗となり、秘密値をBrowser、生成HTML、ログ、fixtureへ出さない。

## TASK-001-02 認証設定とGoogle外部境界を実装する

**目的:** 承認済み設定とGoogle本人確認の外部I/Oを、Applicationから隔離した安全で差し替え可能な境界として提供する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001、NFR-003

**作業内容:** trusted app origin、固定callback、Google OAuth Client、許可メール、DB接続情報、保護鍵をServer限定で扱い、起動時に相互整合を検証する。Googleの認可開始、Token交換、discovery/JWKS、OIDC検証、時刻・乱数・保護に必要な外部能力をApplication portへ提供する。

**受け入れ基準:**

- 環境ごとに単一のtrusted app originと固定callback URIを構成し、必須設定の欠落または不整合時は認証機能を提供せず起動に失敗する。
- Client Secret、許可メール、DB接続情報、保護鍵、tokenおよびcookie値は、必要なServer側外部adapterだけが参照でき、Application入力・出力、Browser、生成HTML、ログ、fixtureへ出ない。
- Google境界は承認済みの固定scope、10秒timeout、認証フロー中retryなし、Server側のToken／OIDC検証契約に従い、CI/E2Eでtest doubleへ差し替えられる。
- `usecase`が必要とする設定、本人確認、時刻・乱数・保護の能力はportで表現され、Google固有型、HTTP応答、環境変数、Secretを`usecase`/`domain`へ渡さない。

## TASK-001-03 管理認証の保護記録とmigrationを実装する

**目的:** OAuth transactionとopaqueな管理sessionを、Server保護記録として安全に保持・失効・保守できるようにする。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015

**作業内容:** 承認済みのmigrationと保護記録の契約を、`usecase` portとPostgreSQL `repository`へ実現する。期限、一回使用、置換、失効、idle期限更新、保守削除および競合時の不変性を維持する。

**受け入れ基準:**

- schema、constraint、index、migrationの初回適用、再適用防止、失敗rollback、中断後の再実行が[DB設計](design/database-schema.md)に一致する。
- `002_admin_session_csrf_ciphertext`は`001_admin_auth_schema`を変更せず、ciphertext列の追加、ciphertextを持たない未失効legacy sessionの失効、適用記録を同一transactionで完了させる。既存sessionのCSRF平文をbackfillせず、migration後は再ログインで新sessionを発行する。
- OAuth transaction、管理session、migration metadataを扱う`usecase` portは、SQL、pool、transaction型を`usecase`/`domain`へ露出しない。
- DBへcookie値、OAuth `state`、CSRF token、PKCE verifierの平文を保存しない。新規管理sessionはCSRF tokenの照合用hashとServer保護ciphertextをともに保存し、ciphertextがNULLのlegacy recordを認可・bootstrapに使わない。
- OAuth transactionは10分・一回使用・置換時無効化、sessionは絶対8時間・idle30分・失効状態を保持し、競合するcallbackでも一度だけ消費される。
- 期限後24時間を超える対象だけを保守削除し、保守削除の失敗を認証許可の理由にしない。

## TASK-001-04 Google OAuth/OIDCによる管理session発行を実装する

**目的:** Google本人確認の開始・callbackを安全に処理し、検証済み許可メールに完全一致するときだけ管理sessionを発行する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001

**作業内容:** OAuth開始とcallbackの`usecase`を実現し、`handler`が承認済みHTTP契約へ変換する。transaction、state、nonce、PKCE、OIDC検証、許可メール照合、条件付き一回使用、session発行の責務を一意に保つ。

**受け入れ基準:**

- OAuth開始は同一originのform POSTだけを受け、成功時にtransaction cookieを設定してGoogleへ`303`する。入力・Origin・保護記録の失敗時はtransactionを変更せず、Googleへ遷移しない。
- callbackはtransactionを原子的に一回使用へ確保してから、Server側だけでToken交換とOIDC検証を行う。
- `email_verified=true`のメールが許可メールと完全一致したときだけsession cookieを発行し、別メール、未検証、失敗、取消、期限切れ、再使用では管理sessionを発行しない。
- callbackのcookie、副作用、redirect、`no-store`、失敗時の秘密情報非露出は[HTTP契約](design/http-contract.md)と[callback操作設計](design/operations/oauth-callback.md)に一致する。
- `handler`は`usecase`の結果をHTTP契約へ変換するだけであり、Google検証・許可メール照合・保護記録操作を直接実行しない。

## TASK-001-05 管理session・CSRF・共通アクセスガードを実装する

**目的:** 後続Featureの管理読取り・状態変更を有効sessionだけに限定し、状態変更をCSRFから保護する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001、NFR-003

**作業内容:** session初期化、管理read guard、管理mutation guard、logoutを、`usecase`の認可規則と`handler`の契約変換へ分離して実現する。cookie、CSRF token、Origin、期限、失効、許可メール再照合、障害時fail closedを承認済み契約どおりに扱う。

**受け入れ基準:**

- session cookieとtransaction cookieの名称、`Secure`、`HttpOnly`、`SameSite`、`Path`、`Domain`、期限、削除属性が[session契約](design/session-contract.md)に一致する。
- session初期化は有効sessionのServer保護ciphertextを復号し、復号値hashと保存hashを定数時間比較してからだけCSRF tokenを返す。無効時は匿名状態へ、読取り・復号・照合障害時はcookie、idle期限、業務状態を変えず`503`へ安全に収束する。
- 管理読取りは有効sessionを必須とし、無効時に業務データを返さない。管理状態変更は有効session、trusted Origin、同期CSRF tokenを必須にし、成功時だけidle期限を更新する。
- logoutは有効sessionでは認可・CSRF後に失効し、trusted Originからの匿名・期限切れ・失効sessionでは業務状態を変更せずcookie削除と匿名状態へ冪等に収束する。Origin不正・CSRF不正・保護記録障害時の副作用は契約どおりとする。
- 公開HTMLの匿名閲覧へ管理認証を要求せず、資格情報を隔離originへ渡さない。

## TASK-001-06 管理認証画面の状態と操作を実装する

**目的:** 管理者が安全にログイン、認証状態の確認、再試行、logoutを行えるようにする。

**関連要件:** REQ-026、REQ-027、BR-015、NFR-001、NFR-003

**作業内容:** Frontendに、同一origin HTTP契約だけを利用する管理認証画面を実装する。初期化、ログイン案内、認証済み、認証基盤障害、callback後の遷移、logoutを、画面設計仕様書の状態・操作・アクセシビリティに対応させる。

**受け入れ基準:**

- 初期化中は管理操作を有効化せず、session bootstrapの結果に応じて認証済み、ログイン案内、認証基盤障害へ遷移する。
- ログイン開始はform POSTと`303` document navigationを使用し、成功・失敗の固定遷移に従う。
- CSRF tokenは画面実行中のメモリだけに保持し、401・403・503・logout成功時の破棄・再取得・再試行が[画面設計仕様書](design/screen-specification.md)に一致する。
- ログイン、再試行、logoutをキーボードで実行でき、状態変化を支援技術へ通知する。小さい画面幅でも主操作と安全な失敗要約を利用できる。
- 許可メール、OAuth値、CSRF token、cookie値、Secret、内部障害詳細を画面・URL・永続ブラウザ保存・通知へ出さない。

## TASK-001-07 認証契約と構造境界の検証を自動化する

**目的:** 分離後のBackend／Frontend境界を含む、認可・セキュリティ・DB・HTTP・画面契約の回帰を検出する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001、NFR-003

**作業内容:** Domain/Application unit、実PostgreSQL integration、Gin HTTP契約、Google test doubleを用いるTLS loopback E2E、Frontend UI、境界・構造検証を[検証責務](design/test-strategy.md)に従って整備する。

**受け入れ基準:**

- Backendが独立Go moduleとして検証でき、`domain`の外部技術非依存、`usecase`のport限定依存、`repository`のport実装と外部I/O隔離、Gin `handler`の契約変換、`cmd`限定の設定読取り・composition、FrontendのHTTP契約限定依存を構造検証で確認する。
- port doubleによるunitで、許可・不許可・未検証・取消・期限切れ・transaction再使用・同時callback、Origin/CSRF、session・logoutの主要な正常・異常系を検証する。
- 実PostgreSQLでmigrationの適用・再適用・rollback、制約・期限・失効・保守削除、保護記録障害時の不変性を検証する。
- `002_admin_session_csrf_ciphertext`の適用・再適用・rollback、ciphertextなしlegacy sessionの失効、新規sessionのCSRF hash/ciphertext保存、ciphertext復号失敗・hash不一致時の`503`と副作用不変を検証する。
- Gin HTTP adapter経由で、HTTPのURI・method・form・303・JSON・status・header・cookie・`no-store`、Origin/CSRF・logout例外を契約どおり検証する。
- TLSを持つloopback trusted app originとGoogle test doubleによるE2Eで、認証画面の遷移、管理読取り拒否、cross-origin拒否、logout分岐、秘密値非露出、`__Host-`/`Secure` cookieを確認する。
- キーボード操作、状態通知、狭い画面幅での主操作を確認する。

## 実装へ委譲する事項

- 承認済みの配置責務を満たす内部file path、package・class・function名、画面component構成
- 承認済みHTTP／DB／画面／境界契約を変えない局所的なAPI選択とコード構造

## 他Featureの責務

- FEAT-002: 生成、検証、再試行、試行履歴とその業務API・画面
- FEAT-003: 版・公開の業務API・画面
- FEAT-004: タグ・掲載場所の業務API・画面
- FEAT-005: 生成HTMLの隔離表示と隔離originの結合検証
