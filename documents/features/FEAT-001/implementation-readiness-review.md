# FEAT-001 実装準備性レビュー

## 対象と独立性

`features/FEAT-001/{requirements.md,design.md,design/**,decisions/**,design-review.md}`、正規要件、Initial Design、承認済みDecision、workflow stateを、設計者および`design-review`担当者とは異なるfresh reviewerとして実装者視点で監査した。今回追加された`design/screen-specification.md`を、HTTP・operation・操作列・検証資料と反証的に照合した。本レビューが変更した成果物は本ファイルだけである。

## 実装者が依存する確定済み契約

- Google OAuth 2.0 Authorization Code FlowとOIDC ID Token検証をServer側で行い、`email_verified=true`のメールとServer限定の許可メールを完全一致で照合する。
- OAuth transactionは10分・一回使用、管理sessionは絶対8時間・アイドル30分である。cookieにはopaqueな参照値だけを置き、PostgreSQLのServer保護記録を正本とする。
- cookie、CSRF header、認証失敗・CSRF失敗・保護記録障害の応答、各認証operationのmethod・URI・入力・成功・失敗はHTTP契約で固定されている。
- 環境ごとの単一trusted app originと固定callback、環境専用OAuth Client、Server限定のSecret、Google test doubleを使う。具体値の配置・Google Console登録はリリース前提条件であり、未設定・不整合では認証endpointを有効化しない。

## 項目別監査

| 項目 | 結果 | 根拠 |
| --- | --- | --- |
| 関係DB・migration | 合格 | `database-schema.md`が`admin_oauth_transactions`と`admin_sessions`のcolumn、型、NULL、主キー・UNIQUE・CHECK・index、失効判定を定義する。`001_admin_auth_schema`の一括transaction、適用記録、再適用防止、失敗rollback、中断後の再実行、期限後24時間の保守削除対象も明記されている。 |
| `oauth_start` wire contract | 合格 | `POST /admin/auth/google/start`、form入力、許可する`return_path`、Origin検査、成功時のGoogle Authorization Endpointへの`303`とtransaction cookie、失敗時のtransaction cookieを発行・置換しない固定`/admin/login?reason=failed`への`303`を`http-contract.md`とoperation資料が定義している。Browser formの失敗をJSONや400/403/503で返さないことも明確である。 |
| `oauth_callback` wire contract | 合格 | `GET /auth/google/callback`、queryとtransaction cookie、成功・失敗時の303先、cookie発行・削除、固定の失敗理由、TokenをURL・表示・ログへ出さない条件が明記されている。 |
| 管理session・read・mutation・logoutのwire contract | 合格 | bootstrapの200 JSON schema、mutation/read guardの401/403/503、CSRF header名・値形式、logoutの有効・匿名・CSRF不正・Origin不正時のcookie副作用が、HTTP契約と各operation資料で一致している。 |
| 管理認証画面 | 合格 | `screen-specification.md`が、`/admin`初期化、ログイン案内、認証済み、認証基盤障害の対象・到達／退出・表示・操作を定義する。各状態は`oauth_start`、callback、bootstrap、read/mutation guard、logoutへ対応付けられ、`200`、`401`、`403`、`503`時のtoken破棄・再ログイン・再初期化・再試行が、操作列を含む関連資料と矛盾しない。callback専用画面を持たず、form POSTから`303` document navigationでGoogleまたは固定失敗案内へ遷移する契約も実装可能である。 |
| 画面の非機能制約 | 合格 | ログイン・再試行・logoutのキーボード操作、状態変化の支援技術向け通知、狭い画面幅での横スクロールなしの主操作、token・cookie・OAuth値・許可メール・Secret・内部障害詳細を画面、URL、永続Browser保存、通知へ出さない制約が明示されている。外観・component構成のみを実装へ委譲しており、安全性・操作性の契約は未決定ではない。 |
| 状態変更・transaction・cleanup | 合格 | 開始時の旧transaction無効化と新規INSERT、callbackの条件付き一回使用確保、Google呼出しをDB transaction外に置くこと、session INSERT失敗時の収束、失効session更新、logout例外、期限後保守削除が定義されている。競合時の更新0件は失敗とし、外部失敗でtransactionを復活させずretryしない。 |
| 設定・秘密値・外部境界 | 合格 | `runtime-configuration.md`がorigin/callbackの構成規則、環境専用Client、Secretと保護鍵のServer限定、設定所有、起動時検証、Google Console同期責任を定義する。Google Token/JWKS/discoveryはServer限定で10秒timeout・retryなし、異常時は認証失敗としてfail closed、CI/E2Eはtest doubleを使う。 |
| 検証責務 | 合格 | `test-strategy.md`がunit、実PostgreSQL integration、HTTP契約、Google test doubleを用いるE2E、隔離境界との責務分担を定義する。migration rollback、競合、cookie、Origin/CSRF、503時の不変性、再使用、秘密値非露出まで対象化されている。 |

## 実装への委譲の確認

file path、symbol、画面コンポーネントの局所構造だけが実装領域へ委譲されている。永続化、wire、設定、外部境界、安全性、検証、画面の状態・操作に関わる契約を「実装時に決める」とする記述は確認されなかった。

`DEC-FEAT-002`で承認された選択肢Aは、具体の本番ドメインをソースコードへ固定するDecisionではなく、環境設定として固定originとcallbackを登録・照合する運用契約である。具体値の未投入は本番リリースを許可しない前提条件であり、実装者が方式を再設計する未解決Decisionではない。

## 残存する実施前提

- 配置運用責任者は、リリース前に環境ごとのorigin・callback URI・Google Console登録・Secretを整合させる。未設定または不一致なら起動時検証で認証endpointを提供しない。
- FEAT-005は、生成HTML隔離originに資格情報が流通しない結合検証を所有する。FEAT-001は信頼済み側から資格情報を渡さない境界テストまでを所有する。

これらは実装仕様の不足ではなく、承認済み設計が定めた配置・Feature境界上の実施条件である。

## 総合判定

**pass**

実装者は、追加の製品・業務・公開・セキュリティ判断をせずにFEAT-001へ着手できる。DDL/migration、operation別wire contract、状態変更と外部呼出しのtransaction・cleanup、設定とGoogle境界、テスト責務に加え、管理認証画面の状態・操作・API対応・エラー動作・アクセシビリティ・responsive・秘密情報非露出が設計資料間で追跡可能であり、未解決のL3/L4事項はない。

## 次ゲート

利用者による詳細設計レビュー・承認。その承認後にだけTask分割へ進む。
