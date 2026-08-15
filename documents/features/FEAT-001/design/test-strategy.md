# 管理認証の検証責務

| 層 | 確認する契約 |
| --- | --- |
| unit | `domain`/`usecase`の設定形式、相対復帰先、乱数・hash比較、期限、許可メール照合、Origin/CSRF判定、OIDC claim検証結果の分類。port doubleを使い、Gin、PostgreSQL、Google、環境設定、`repository`実装へ接続しない。 |
| integration | 実PostgreSQLで、`schema_migrations`の初回DDL、`001_admin_auth_schema`の適用記録が対象DDL成功後だけに作成されること、再適用防止、失敗時のmetadata/対象DDL/適用記録のrollback、unique制約、transaction一回使用の競合、session失効・期限、callbackの条件付き消費とsession保存、保守削除を確認する。 |
| HTTP契約 | 各operationのmethod/URI/status/JSON/header/cookie、`no-store`、cookie属性・削除、401/403、保護記録障害時の503とcookie・業務状態不変、`oauth_start`のOrigin・form field不正・保護記録不可時にtransactionを発行・置換せず固定失敗案内へ303すること、成功時の303がGoogleへのdocument navigationになること、Origin不正logout時にcookieを削除しないこと。 |
| E2E | TLSを持つloopback trusted app originで、Google test double経由の許可・不許可・取消・期限切れ・再使用、`__Host-`/`Secure` transaction・session cookieを伴う同一originのOAuth開始・CSRF成功、cross-origin拒否、管理読取りの未認可拒否、logoutの有効／匿名分岐を確認する。loopback HTTPは設定検証の対象にできるが、このBrowser OAuth E2Eの前提にはしない。 |
| 境界 | 認可情報・token・Secret・許可メールをHTML、URL、ログ、fixture、生成HTMLへ出さないこと。本格的な隔離origin結合検証はFEAT-005が所有する。 |
| 構造 | `backend/`が独立Go moduleであり、Go依存を`backend/go.mod`/`backend/go.sum`だけで管理すること。`cmd`、`handler`、`usecase`、`repository`、`domain`を明示的に配置し、環境変数・Secret storeを直接読むのが`cmd`だけであること、`domain`が他層へ、`usecase`がGin・PostgreSQL・Google・環境設定・`repository`実装へ、`handler`が`repository`へ、`repository`が`handler`へ直接依存しないことを確認する。Gin `handler`のrequest/response/cookie/redirect変換はHTTP契約テストで確認する。`frontend/`がBackend内部・DB・Secretを参照せず、HTTP契約だけを使うこと。 |

fixtureは架空のClient、鍵、メール、tokenだけを使い、実運用値を含めない。
