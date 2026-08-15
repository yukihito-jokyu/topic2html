# FEAT-001 実装準備性レビュー

## 対象と独立性

`requirements.md`、`design.md`、`design/**`、`decisions/**`、`design-review.md`、正規要件、Initial Design、承認済みDecisionを、設計者およびdesign-review担当者とは異なる実装者視点で監査した。本レビューが変更する成果物は本ファイルだけである。

監査対象は、利用者が明示したBackendの`handler`／`usecase`／`repository`／`domain`分離と、既存のOAuth、PostgreSQL、HTTP、session、CSRF、Frontend、秘密情報境界を後退させずに構造移行できるかである。

## 実装者が依存できる確定済み契約

- `backend/`は独立Go moduleであり、`cmd`、`handler`、`usecase`、`repository`、`domain`を明示配置する。`cmd`だけが環境変数・Secret storeから全Server設定を読み、存在・形式・origin/callback整合を検証して、各層をconstructorで組み立てる。
- `handler`はGinのroute、middleware、request/response、cookie、redirect変換だけを担い、`usecase`を呼ぶ。`usecase`は`domain`と自ら定義するportだけに依存し、repository実装、Gin、SQL、Google、環境設定を参照しない。`repository`はportを実装して外部I/Oを閉じ込め、環境変数・Secret storeを直接読まない。`domain`は外部層に依存しない。
- Google OAuth 2.0 Authorization Code FlowとOIDC ID TokenをServer側で検証し、`email_verified=true`のメールをServer限定の許可メール1件と完全一致で照合する。Authorization Requestのscopeは`openid email`に固定する。
- OAuth transactionは10分・一回使用、管理sessionは絶対8時間・idle 30分であり、opaque参照値と同期CSRF tokenをServer保護記録に保持する。HTTP operation、cookie、CSRF header、認証／CSRF／保護記録障害のwire形式、画面状態・操作が定義済みである。
- `schema_migrations`、`admin_oauth_transactions`、`admin_sessions`のDDL相当、`001_admin_auth_schema`のtransaction内での作成・version記録・rollback・再適用判定、および実PostgreSQL検証の期待結果が定義済みである。
- Browser OAuth E2EはTLSを持つloopback trusted app originとGoogle test doubleを使用し、`__Host-`/`Secure` cookieを緩和しない。HTTP loopbackは設定検証に限る。

## 項目別監査

| 項目 | 結果 | 根拠・確認内容 |
| --- | --- | --- |
| 4層・composition root | 合格 | `DEC-ARCH-003`、`design.md`、`architecture-boundaries.md`、`test-strategy.md`が物理配置、責務、許可依存・禁止依存を一致して定める。`cmd`だけが設定読取り・起動時検証・constructor注入を行い、repositoryは注入済みのSecretを使うだけで直接読取りをしない。 |
| 構造検証 | 合格 | `domain`の外部依存、`usecase`のGin／PostgreSQL／Google／環境設定／repository実装依存、`handler → repository`、`repository → handler`、`cmd`以外の設定・Secret読取りを検出する契約がある。Gin handlerのwire変換はHTTP契約テストで検証する。 |
| HTTP／画面・Gin境界 | 合格 | `http-contract.md`、operation資料、`screen-specification.md`が、method、URI、入力、status、JSON、redirect、cookie、header、副作用、初期化、ログイン、callback後、logout、401/403/503、アクセシビリティ、responsive、秘密情報非露出を対応付ける。handlerがHTTP固有の変換だけを担当するため、usecaseへGin型を渡さない実装が可能である。 |
| OAuth、session、CSRF、cleanup | 合格 | `session-contract.md`とoperation資料が期限、一回使用、条件付き消費、Origin/CSRF、idle更新、logout例外、Google呼出しをDB transaction外に置くこと、失敗時にtransaction/sessionを復活させないことを定める。 |
| PostgreSQL・migration・repository | 合格 | `database-schema.md`がmetadata table、migration version、DDL／記録順、失敗時rollback、再適用防止、保護記録二表のcolumn・制約・index、operation別transaction/access mapを定める。SQL・pool・transaction型をusecaseへ出さないport契約と整合する。 |
| 設定・秘密値・Google境界 | 合格 | `runtime-configuration.md`がcmdによるfail-closed検証、固定callback、constructor注入、10秒timeout、retryなし、test double、秘密値非露出を定める。Client Secret・保護鍵・DB接続情報をusecaseの公開入出力へ入れないため、層境界と安全要件が矛盾しない。 |
| Frontend・外部境界 | 合格 | Frontendは同一originのHTTP契約だけに依存し、Backend内部、DB、Google、Secretを参照しない。生成HTMLの別origin結合検証はFEAT-005へ明確に分離されている。 |
| 検証責務 | 合格 | unit、実PostgreSQL integration、Gin HTTP契約、TLS loopbackのGoogle test double E2E、境界・構造検証の分担が具体化され、fixtureの秘密値非露出も定められている。 |

## 実装への委譲の検出

package、file、symbol、画面component、構造検査の具体的な実装手段だけが実装領域へ適切に委譲されている。一方、4層の物理配置・責務・依存方向、設定の所有、HTTP wire、cookie属性、CSRF、DB transaction・cleanup、外部通信、TLS E2E前提は確定済みである。挙動・永続化・wire・設定・安全性を実装者が再設計すべき箇所は検出されなかった。

## 残存リスク

- TLS loopbackの証明書発行・信頼方法、test doubleの具体的な起動方法、構造検査の具体的手段は実装選択である。`__Host-`/`Secure` cookieを緩和せず、指定のE2Eと構造検査を実行して受け入れる。
- Google Consoleへの環境別origin／callback登録とSecret注入はDEC-FEAT-002どおり配置運用責任者のリリース前提である。通常CIは実Google資格情報を用いない。

## 総合判定

**pass**

`handler`／`usecase`／`repository`／`domain`と`cmd`の分離は、設定・Secretの所有、依存方向、構造検証まで実装可能な粒度で確定している。Gin HTTP変換、OAuth、PostgreSQL、migration、session、CSRF、Frontend、秘密情報境界の契約もこの分離と整合し、構造移行のために新たな製品・業務・安全性判断をする必要はない。

## 次ゲート

利用者による更新後の詳細設計の明示承認。承認前にTask分割・implementation handoffへ進んではならない。承認後は`task-breakdown`へ委譲する。
