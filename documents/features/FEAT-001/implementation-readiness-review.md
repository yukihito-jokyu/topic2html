# FEAT-001 実装準備性レビュー

## 対象と独立性

`requirements.md`、`design.md`、`design/**`、`decisions/**`、`design-review.md`、正規要件、Initial Design、承認済みDecision、ならびに現行`backend/`を、設計者・design-review担当者・過去のpreflight/実装担当とは異なる実装者視点で監査した。本レビューが変更する成果物は本ファイルだけである。

重点は、承認済みDEC-FEAT-004によるCSRF ciphertext migration、既存sessionの失効、bootstrap／管理ガードのHTTP・永続化契約、現行Backendへの接続可能性である。

## 実装者が依存できる確定済み契約

- `002_admin_session_csrf_ciphertext`は`001_admin_auth_schema`の後に、各migrationを個別transactionで昇順適用する。`002`はNULL可の`csrf_token_ciphertext BYTEA`を追加し、ciphertextを持たない未失効legacy sessionを同一transaction内で失効する。DDL、失効更新、適用記録のいずれかが失敗すれば全てrollbackし、記録済みversionは再適用しない。
- 新規sessionは256 bit CSRF tokenのhashと、同じtokenをServer保護鍵で暗号化したciphertextを保存する。平文はDB、cookie、redirect、URL、ログ、fixture、永続Browser保存へ出さない。ciphertextがNULLのlegacy recordは認可・bootstrapに使わない。
- bootstrapは有効sessionに限りciphertextをServer内で復号し、復号値のhashと保存hashを定数時間比較してから、`Cache-Control: no-store`の`200 {"authenticated":true,"csrf_token":"..."}`で返す。DB読取り・復号・hash照合の失敗は、cookie、idle期限、業務状態を変更せず`503 authentication_unavailable`へfail closedする。
- read guardは有効sessionと許可メール再照合を要求し、無効なら業務データを返さず`401`へ収束する。mutation guardはさらにtrusted `Origin`と`X-CSRF-Token`を要求し、全検査成功時だけ後続業務更新と同一transactionでidle期限を更新する。logoutだけはtrusted Origin上の匿名・無効sessionをcookie削除と匿名応答へ冪等に収束させる。
- `cmd`だけが設定・Secretを読取り検証し、`handler`はGinでHTTP契約へ変換、`usecase`はdomainとportだけ、`repository`はPostgreSQL・Google・保護機能の外部I/Oだけを担う。trusted origin、callback、保護鍵、Google timeout/retry、test double、TLS loopback E2Eの運用責任も定義済みである。

## 現行Backendとの接続監査

| 現行証拠 | 更新後に実装する接続 | 判定 |
| --- | --- | --- |
| `backend/repository/postgres/migrations.go`は`001`をtransaction・`schema_migrations`で冪等適用し、`cmd/migrate`がこれを呼ぶ。 | 同じrunnerに`002`を昇順で追加する。`001`は変更・再適用せず、`002`で列追加とlegacy session失効を一つのtransactionにする。 | 合格。migrationのversion、順序、失敗時の挙動が設計で固定されている。 |
| `domain/auth.AdminSession`、PostgreSQL `CreateAdminSession`／`FindAdminSession`、OAuth callbackは現時点でCSRF hashだけを扱う。 | domain record、repository SQL、callbackのsession保存をhashとciphertextの両方を扱うよう拡張し、新規recordでciphertextを必須にする。 | 合格。既存の責務境界を越えず、追加するfieldと保存・読出しの契約が固定されている。 |
| `repository/security.Service`はServer限定AES-256-GCMの`Seal`／`Open`を提供し、OAuth PKCE verifierで使用済みである。 | 同じ注入済み保護portでCSRF tokenをseal/openし、bootstrapでhashを再計算して定数時間照合する。 | 合格。保護鍵の新設、Browser露出、別の暗号方式の採用を必要としない。 |
| 現行routerはOAuth開始・callbackだけを公開し、session bootstrap、logout、read/mutation guardは未実装である。 | `handler`が確定済みHTTP contractのsession・logout routeと、後続Featureで利用するread/mutation guardへ変換する。認可・CSRF・永続化の判定はusecase/portへ置く。 | 合格。未実装部分はTASK-001-05の対象であり、wire・副作用・失敗応答が定義済みである。 |

## 項目別監査

| 項目 | 結果 | 根拠・確認内容 |
| --- | --- | --- |
| DB schema・migration | 合格 | `database-schema.md`がtable、column、型、NULL、key、index、`001`/`002`の順序、適用記録、rollback、legacy session失効、保守削除を定義する。migration時の再ログインというL3影響はDEC-FEAT-004で承認済みである。 |
| CSRF ciphertext + hash | 合格 | `session-contract.md`、callback／bootstrap operation、DB資料が、生成、seal、保存、open、hash再計算、定数時間比較、503 fail closed、平文非露出を一貫して定める。ciphertext不在を匿名成功やtoken再発行へ変換する余地はない。 |
| HTTP・operation | 合格 | `http-contract.md`とoperation資料がbootstrap、logout、read guard、mutation guardのmethod、URI、入力、status、JSON、cookie、`no-store`、Origin/CSRF、業務・cookie副作用を定義する。logoutの匿名例外も共通guardと明示的に整合している。 |
| transaction・競合・cleanup | 合格 | callbackのtransaction一回使用は短いtransaction、Google呼出しはその外、新session insertは別transactionとする。mutationのidle更新は業務更新と同一transaction、migrationのlegacy失効はDDL・version記録と同一transaction、保守失敗は認証許可の理由にしないことが定義済みである。 |
| 設定・外部境界 | 合格 | `runtime-configuration.md`がcmd限定の読取り・起動時検証、保護鍵の注入、固定callback、trusted origin、10秒timeout・retryなし、通常CI/E2EのGoogle test doubleを定義する。実装者が公開・秘密値・運用責任を再決定する必要はない。 |
| 画面・外部origin | 合格 | `screen-specification.md`はbootstrap tokenの実行時メモリ保持、401/403/503時の破棄・再取得・再試行、logout状態を定義する。管理資格を生成HTML隔離originへ流通させず、隔離origin結合検証をFEAT-005へ分離している。 |
| 検証戦略 | 合格 | `test-strategy.md`はciphertext復号失敗・hash不一致、`002`の適用／rollback／legacy失効、新session保存、503不変性、Origin/CSRF、TLS loopback E2Eをunit、実PostgreSQL integration、HTTP、E2Eへ割り当てる。実Googleや実運用Secretを通常CIで用いない。 |

## 実装への委譲の検出

migration SQLの局所的な記述、Goのfile/package/symbol、port・test double・TLS証明書の具体的構成だけが実装へ委譲されている。これらは承認済みのDB、HTTP、security、設定、検証契約を変えない実装選択である。CSRF平文の扱い、legacy sessionの互換方針、migration原子性、認可・CSRF失敗時の外部応答・副作用、鍵の所有など、実装者が再設計すべき製品・公開・security・architecture判断は検出されなかった。

## workflow整合に関する注意

本レビュー時点の`documents/.ai/workflow/state.yaml`は`design_review`と`implementation_readiness_review`を`pending`、`pending_decisions`のDEC-FEAT-004を`pending_human_approval`としている一方、DEC-FEAT-004本文は2026-08-15に承認済みであり、`design-review.md`は`pass`である。また既存handoffには`implementation_ready`が先行記載されている。本レビューは設計契約の可否だけを判定する。stateとhandoffの正本整合はplanning-orchestratorが後続gateの完了に応じて更新する必要があり、実装者が補正してはならない。

## 総合判定

**pass**

承認済みDEC-FEAT-004を含め、DB/migration、CSRF ciphertextとhash、HTTP/operation、runtime configuration、test strategy、現行Backendへの接続が実装可能な粒度で定義されている。TASK-001-05の実装者は、新たなsecurity・公開契約・architecture判断を行わずに着手できる。

## 次ゲート

利用者による更新後の詳細設計の明示承認。承認済みであれば、planning-orchestratorはworkflow stateとhandoffの整合を回復した上でtask-breakdown以降の正規手順へ進める。
