# FEAT-001 Clean Architecture境界と移行契約

## 目的

既存の認証設定・Google境界を後退させず、FEAT-001を`backend/`と`frontend/`へ分離し、Ginを`handler`に限定して実装・移行できるようにする。本資料は[詳細設計](../design.md)の責務を補い、4層とcomposition rootの物理配置以外の具体的なfile名・package名・関数名を固定しない。

## 配置責務

| 配置 | 含めるもの | 含めないもの |
| --- | --- | --- |
| `backend/` | `cmd`、`handler`、`usecase`、`repository`、`domain`、Migration、Backend検証。 | React UI資産、生成HTML、Browserへ配布するSecret。 |
| `frontend/` | 管理UIのNode.js/TypeScript/React資産、Frontendの型検査・unit/UI検証。 | Go内部package、PostgreSQL接続、Google Client Secret、session/CSRF/OAuth値の永続保存。 |
| 生成HTML隔離origin | FEAT-005が所有する非信頼表示の配置・実行境界。 | FEAT-001の管理session、管理API client、Google/Codex/DB資格情報。 |

Backendの物理ディレクトリは、少なくとも`cmd`、`handler`、`usecase`、`repository`、`domain`を区別して依存を可視化する。`cmd`だけが全Server設定・Secretを読取り起動時に検証し、検証済み値を各層へconstructor注入する。PostgreSQL、Google、暗号、migrationの外部I/O実装は`repository`に置くが、`repository`は環境変数・Secret storeを直接読まない。承認済み[DEC-FEAT-003](../decisions/DEC-FEAT-003.md)により、`backend/`は独立Go module・独立検証単位であり、Go依存の正本は`backend/go.mod`と`backend/go.sum`だけとする。

## FEAT-001の責務対応

| 操作・横断責務 | `domain` | `usecase` | `handler` | `repository` |
| --- | --- | --- | --- | --- |
| `oauth_start` | transaction期限・一回使用に必要な値規則 | `cmd`注入済み認可依存で入力を正規化し、既存transaction無効化、新規transaction発行 | form/OriginをHTTP契約どおりに読み、303/cookieへ変換 | `cmd`注入済みGoogle設定・秘密値、乱数・hash・暗号、PostgreSQL、Google Authorization URL生成 |
| `oauth_callback` | transaction/session有効性、期限、資格状態 | `cmd`注入済み認可依存でtransaction消費、Google検証結果と許可メール照合、session作成 | query/cookieを受け、成功/失敗303とcookie削除を変換 | PostgreSQL、`cmd`注入済みGoogle Token/OIDC設定・Secret、乱数・hash・暗号、時刻 |
| session bootstrap | 有効sessionの判定、CSRF token復元規則 | `cmd`注入済み認可依存でsession照会とServer保護ciphertextの復号を行い、匿名/認証済み結果を作成 | cookieからHTTP JSON/cookie削除へ変換 | PostgreSQL、`cmd`注入済み保護鍵、時刻、hash比較 |
| read/mutation guard | 認可済み資格・期限の規則 | `cmd`注入済み認可依存でsession/許可メール/CSRF/Originを検査、成功時の許可結果 | Gin middlewareまたはroute境界でHTTP入力を`usecase`へ渡す | PostgreSQL、`cmd`注入済み保護鍵、時刻、hash比較 |
| logout | session失効と匿名logout例外の規則 | `cmd`注入済み認可依存でOrigin/CSRF条件の評価と失効 | POSTをHTTP結果/cookie削除へ変換 | PostgreSQL、`cmd`注入済み保護鍵、時刻、hash比較 |

## 移行の不変条件

1. 現在の[HTTP契約](http-contract.md)、[session契約](session-contract.md)、[DB schema](database-schema.md)、[設定・Google境界](runtime-configuration.md)を変更せず、担当層だけを移す。
2. Ginは`handler`のHTTP request/response、route、middleware、cookie、redirectにだけ使う。`usecase`/`domain`の公開モデルにGin型を入れない。
3. PostgreSQLのSQL・pool・transaction型、GoogleのHTTP/SDK型、環境変数・Secret storeの読取りを`usecase`/`domain`へ持ち込まない。外部I/Oは`repository`だけが実装するが、環境変数・Secret storeの読取りと起動時検証は`cmd`だけが担い、必要な値を`repository`へconstructor注入する。
4. Frontendは[画面設計仕様書](screen-specification.md)とHTTP契約に従い、認可判定を再実装しない。認可・CSRFの最終判断はBackendだけが行う。
5. 既存の初期Go Serverから移す際、設定不備では起動しない、Google通信10秒timeout・retryなし、test double注入、秘密値非露出という既存安全契約を維持する。

## 構造検証

- Backendの静的検査または依存検査で、環境変数・Secret storeを直接読むのが`cmd`だけであること、`domain`が他層へ、`usecase`がGin、PostgreSQL、Google、環境設定、`repository`実装へ、`handler`が`repository`へ、`repository`が`handler`へ直接importしないことを確認する。
- `usecase` unit testはport doubleで、OAuth開始/callback、session、guard、logoutの正常・主要異常系を確認する。
- HTTP契約テストはGin `handler`経由で、method/URI/status/JSON/header/cookie/303/no-storeを確認する。
- PostgreSQL/Googleのintegration testは`repository`と`usecase` portの接続を確認し、通常CIのGoogle境界にはtest doubleを使用する。
- Frontend検証はHTTP契約に対する画面状態だけを確認し、Backend内部・DB・Secretへの依存がないことを確認する。
- Go Module検証、整形、静的検査、unit/integration test、buildは`backend/`を作業ディレクトリとして実行する。root Go moduleにBackend依存を残さず、Frontendとの横断確認はHTTP契約とE2Eで行う。
