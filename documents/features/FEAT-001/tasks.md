# FEAT-001 実装タスク — 管理者本人確認と管理アクセス

## Task一覧

| ID | タイトル | 論理領域 | 依存 |
| --- | --- | --- | --- |
| TASK-001-01 | 認証設定と起動時検証を整備する | 設定・外部境界 | なし |
| TASK-001-02 | 管理認証の保護記録とmigrationを実装する | 永続化・migration | TASK-001-01 |
| TASK-001-03 | Google OAuth/OIDCによる管理session発行を実装する | 認証・外部連携 | TASK-001-01、TASK-001-02 |
| TASK-001-04 | 管理session・CSRF・共通アクセスガードを実装する | 認可・セキュリティ | TASK-001-01、TASK-001-02 |
| TASK-001-05 | 管理認証画面の状態と操作を実装する | presentation | TASK-001-03、TASK-001-04 |
| TASK-001-06 | 認証契約の検証を自動化する | testing / verification | TASK-001-01〜05 |

## TASK-001-01 認証設定と起動時検証を整備する

**目的:** 環境ごとに固定したtrusted app origin、callback URI、Google OAuth Client、許可メール、DB接続情報、保護鍵をServer限定で安全に利用できるようにする。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001、NFR-003

**作業内容:** 設定の所有・非露出、originとcallbackの構成規則、起動時検証、Google境界のtimeout・retryなし・test double利用を、承認済み契約どおりに実現する。

**受け入れ基準:**

- 環境ごとに単一のtrusted app originと固定callback URIを構成し、起動時に形式と相互整合を検証する。
- Google Client Secret、許可メール、DB接続情報、保護鍵はServer実行環境だけで参照でき、Browser・生成HTML・ログ・fixtureへ出ない。
- 必須設定の欠落または不整合時は、管理認証endpointを提供せず起動に失敗する。
- Googleのdiscovery/JWKS/Token境界は10秒timeout、認証フロー中retryなしとし、CI/E2Eではtest doubleを利用できる。

## TASK-001-02 管理認証の保護記録とmigrationを実装する

**目的:** OAuth transactionとopaqueな管理sessionをServer保護記録として安全に保持・失効・保守できるようにする。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015

**作業内容:** `001_admin_auth_schema`、OAuth transaction、管理session、期限・一回使用・失効・保守削除、およびoperation別のDB access mapを実現する。

**受け入れ基準:**

- schema・constraint・index・migrationの適用、再適用防止、失敗rollback、中断後の再実行が[DB設計](design/database-schema.md)に一致する。
- DBへcookie値、OAuth `state`、CSRF token、PKCE verifierの平文を保存しない。
- OAuth transactionは10分、一回使用、置換時無効化となり、sessionは絶対8時間・idle30分・失効状態を保持する。
- 期限後24時間を超える対象だけを保守削除し、保守削除の失敗を認証許可の理由にしない。

## TASK-001-03 Google OAuth/OIDCによる管理session発行を実装する

**目的:** Google本人確認の開始・callbackを安全に処理し、許可メールに一致したときだけ管理sessionを発行する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001

**作業内容:** `oauth_start`と`oauth_callback`を、state・nonce・PKCE・ID Token検証・許可メール完全一致・transaction一回使用の契約に従って実現する。

**受け入れ基準:**

- `POST /admin/auth/google/start`は同一originのform POSTだけを受け、成功時にtransaction cookieを設定してGoogleへ`303`する。
- 入力・Origin・保護記録の失敗時はtransactionを変更せず、固定の失敗案内へ`303`し、Googleへ遷移しない。
- callbackはtransactionを原子的に一回使用へ確保し、Google Token交換とOIDC検証をServer側だけで行う。
- `email_verified=true`のメールが許可メールと完全一致したときだけsession cookieを発行し、それ以外の成功・失敗・取消・再使用では管理sessionを発行しない。
- callbackの成功・失敗のcookie、副作用、redirect、`no-store`が[HTTP契約](design/http-contract.md)と[callback操作設計](design/operations/oauth-callback.md)に一致する。

## TASK-001-04 管理session・CSRF・共通アクセスガードを実装する

**目的:** 後続Featureの管理読取り・状態変更を有効sessionだけに限定し、状態変更をCSRFから保護する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001、NFR-003

**作業内容:** session初期化、管理read guard、管理mutation guard、logout、cookie属性、CSRF token、失効・期限・許可メール再照合、fail closedを実現する。

**受け入れ基準:**

- session cookieとtransaction cookieは指定された`__Host-`名、`Secure`、`HttpOnly`、`SameSite=Lax`、`Path=/`、`Domain`未指定、期限・削除属性を守る。
- `GET /admin/auth/session`は有効sessionだけにCSRF tokenを返し、無効時は匿名状態へ、保護記録障害時は`503`へ安全に収束する。
- 管理読取りは有効sessionを必須とし、無効時に業務データを返さない。管理状態変更は有効session、trusted Origin、同期CSRF tokenを必須にし、成功時だけidle期限を更新する。
- logoutは有効sessionでは認可・CSRF後に失効し、trusted Originからの匿名・期限切れ・失効sessionでは業務状態を変えずcookie削除と匿名状態へ冪等に収束する。Origin不正・CSRF不正・保護記録障害時のcookie副作用は契約どおりとする。
- 公開HTMLの匿名閲覧へ管理認証を要求せず、資格情報を隔離originへ渡さない。

## TASK-001-05 管理認証画面の状態と操作を実装する

**目的:** 管理者が安全にログイン、認証状態の確認、再試行、logoutを行えるようにする。

**関連要件:** REQ-026、REQ-027、BR-015、NFR-001、NFR-003

**作業内容:** `/admin`初期化、ログイン案内、認証済み、認証基盤障害の画面状態を、認証APIおよび後続Featureの共通ガードへ接続する。

**受け入れ基準:**

- 初期化中は管理操作を有効化せず、session bootstrapの結果に応じて認証済み、ログイン案内、認証基盤障害へ遷移する。
- ログイン開始はform POSTと`303` document navigationを使用し、成功・失敗の固定遷移に従う。
- CSRF tokenは画面実行中のメモリだけに保持し、401・403・503・logout成功時の破棄・再取得・再試行が[画面設計仕様書](design/screen-specification.md)に一致する。
- ログイン、再試行、logoutをキーボードで実行でき、状態変化を支援技術へ通知する。小さい画面幅でも主操作と安全な失敗要約を利用できる。
- 許可メール、OAuth値、CSRF token、cookie値、Secret、内部障害詳細を画面・URL・永続ブラウザ保存・通知へ出さない。

## TASK-001-06 認証契約の検証を自動化する

**目的:** 認可・セキュリティ・DB・HTTP・画面契約の回帰を検出する。

**関連要件:** REQ-026、REQ-027、BR-001、BR-015、NFR-001、NFR-003

**作業内容:** unit、実PostgreSQL integration、HTTP契約、Google test doubleを用いるE2E、画面アクセシビリティ・境界テストを整備する。

**受け入れ基準:**

- 許可・不許可・未検証・取消・期限切れ・transaction再使用・同時callbackを検証する。
- migrationの適用・再適用・rollback、制約・期限・失効・保守削除、保護記録障害時の不変性を実PostgreSQLで検証する。
- HTTPのURI・method・form・303・JSON・status・header・cookie・`no-store`、Origin/CSRF・logout例外を契約どおり検証する。
- Google test doubleによるE2Eで、認証画面の遷移、管理読取り拒否、cross-origin拒否、logout分岐、秘密値非露出を確認する。
- キーボード操作、状態通知、狭い画面幅での主操作を確認する。

## 実装へ委譲する事項

- file path、package・class・function名、画面component構成、framework固有API
- 承認済み契約を変えない範囲での外観・局所的なコード構造

## 他Featureの責務

- FEAT-002: 生成、検証、再試行、試行履歴とその業務API・画面
- FEAT-003: 版・公開の業務API・画面
- FEAT-004: タグ・掲載場所の業務API・画面
- FEAT-005: 生成HTMLの隔離表示と隔離originの結合検証
