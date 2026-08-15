# 管理認証の検証責務

| 層 | 確認する契約 |
| --- | --- |
| unit | 設定形式、相対復帰先、乱数・hash比較、期限、許可メール照合、Origin/CSRF判定、OIDC claim検証結果の分類。 |
| integration | Migrationの適用・再適用・rollback、unique制約、transaction一回使用の競合、session失効・期限、callbackの条件付き消費とsession保存、保守削除。実PostgreSQLを使う。 |
| HTTP契約 | 各operationのmethod/URI/status/JSON/header/cookie、`no-store`、cookie属性・削除、401/403、保護記録障害時の503とcookie・業務状態不変、`oauth_start`のOrigin・form field不正・保護記録不可時にtransactionを発行・置換せず固定失敗案内へ303すること、成功時の303がGoogleへのdocument navigationになること、Origin不正logout時にcookieを削除しないこと。 |
| E2E | Google test double経由の許可・不許可・取消・期限切れ・再使用、同一originのOAuth開始・CSRF成功、cross-origin拒否、管理読取りの未認可拒否、logoutの有効／匿名分岐。 |
| 境界 | 認可情報・token・Secret・許可メールをHTML、URL、ログ、fixture、生成HTMLへ出さないこと。本格的な隔離origin結合検証はFEAT-005が所有する。 |

fixtureは架空のClient、鍵、メール、tokenだけを使い、実運用値を含めない。
