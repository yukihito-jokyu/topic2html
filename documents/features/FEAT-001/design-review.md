# FEAT-001 詳細設計レビュー

## 対象と独立性

`features/FEAT-001/{requirements.md,design.md,design/**,decisions/**}`を、設計者および過去のレビュー・実装担当の結論に依存せず、正規要件、Initial Design、承認済みDecision、既存のmigration／保護record実装を反証的に照合した。今回の重点は、CSRF tokenのServer保護ciphertext、`002_admin_session_csrf_ciphertext`、ciphertextを持たないlegacy管理sessionの失効である。

## 正規ソース・Decision台帳

| 優先度 | ソース | 状態 | 適用範囲・確定内容 |
| --- | --- | --- | --- |
| 1 | `requirements/requirements.md` REQ-026–027、`business-rules.md` BR-001、BR-015–016 | confirmed | 許可メール1件の利用者だけに管理操作を許可し、管理情報・資格情報・秘密値を閲覧者または失敗表示へ出さない。 |
| 1 | `requirements/nonfunctional-requirements.md` NFR-001、NFR-003 | confirmed | 認証情報、session、管理操作・管理情報を生成HTMLへ渡さない。 |
| 2 | `design/architecture.md`、`cross-cutting-concerns.md`、`DEC-ARCH-003.md` | 承認済み L3 | 信頼済みServerが認証・認可を所有し、GinはHTTP境界だけで使う。PostgreSQL migrationは版管理し、保護値をBrowser・ログ・fixtureへ出さない。 |
| 2 | `features/FEAT-001/decisions/DEC-FEAT-001.md` | 承認済み L3 | Server側永続session、同期CSRF token、状態変更時の厳格なOrigin照合を採用する。 |
| 2 | `features/FEAT-001/decisions/DEC-FEAT-002.md`、`DEC-FEAT-003.md` | 承認済み L3 | 単一trusted originとServer限定Secret、`backend/`独立Go moduleを採用する。 |
| 2 | `features/FEAT-001/decisions/DEC-FEAT-004.md` | 承認済み L3（2026-08-15） | `002_admin_session_csrf_ciphertext`で復元用ciphertextを追加し、ciphertextを持たない未失効legacy sessionを同一transactionで失効する。既存利用者には再ログインを求める。 |
| 3 | `features/FEAT-001/design.md`、`design/**` | 詳細設計 | 上記の確定事項をHTTP、DB、operation、画面、設定、検証へ具体化する。 |

## 契約根拠の追跡

| 境界 | 設計上の契約 | 正規根拠 | 判定 | レビュー結果 |
| --- | --- | --- | --- | --- |
| 本人確認・認可 | Server側で検証済みの`email_verified=true`メールを許可メール1件と完全一致で照合する。 | REQ-026–027、BR-001、DEC-ARCH-003 | explicit | callback、session、設定資料で一貫する。 |
| CSRF発行・保存 | callbackが256 bit CSRF tokenを生成し、hashとServer保護鍵によるciphertextだけを新規sessionへ保存する。平文をDB、cookie、redirect、ログへ出さない。 | DEC-FEAT-001、DEC-FEAT-004、BR-015–016 | explicit | `session-contract.md`、callback operation、DB schema、設定資料、test strategyが一致する。 |
| CSRF復元 | bootstrapは有効sessionに限りciphertextをServer内で復号し、復号値をhash化して保存hashと定数時間比較し、一致時だけ`no-store`の同一origin応答で平文を返す。 | DEC-FEAT-001、DEC-FEAT-004 | derived | hashは平文を復元できないため、承認済みのciphertext保存が復元経路を一意に補う。復号失敗・hash不一致はtokenを返さず503でfail closedと定める。 |
| CSRF状態変更照合 | mutation／有効session logoutはtrusted `Origin`とbootstrap由来tokenのhash照合をともに要求する。 | DEC-FEAT-001 | explicit | operation、HTTP、画面仕様、test strategyで一致する。 |
| legacy session移行 | `001`を変更せず、`002`がNULL可ciphertext列を追加し、NULLかつ未失効のsessionを同じmigration transactionで失効する。 | DEC-FEAT-004 | explicit | DB資料がversion、順序、同一transaction、再ログインを定義する。既存ciphertextの安全なbackfillを試みない。 |
| migration後の認可 | NULL ciphertextのlegacy recordは失効済みであり、bootstrap／管理guardで匿名・再ログインへ収束する。新規sessionはciphertextを必須とする。 | DEC-FEAT-004 | explicit | session、DB、bootstrap、read/mutation guard、HTTP、E2Eの契約が矛盾しない。 |
| migration原子性・再適用 | metadata確認、対象DDL、legacy session失効、version記録を一つのtransactionで行い、失敗時は全てrollbackし、記録済みversionを再適用しない。 | DEC-ARCH-003、DEC-FEAT-004 | derived | 既存のmigration runner契約およびDB資料の順序と一致する。部分適用でlegacy sessionを誤って認可可能にする経路はない。 |
| 保護記録障害 | DB読取り、ciphertext復号、hash照合に失敗した場合は503 `authentication_unavailable`、cookie・idle期限・業務状態は不変とする。 | DEC-FEAT-001、DEC-FEAT-004 | derived | 安全に利用できない保護記録ではfail closedとする承認済み方針を具体化している。 |
| 秘密情報境界 | CSRF平文はbootstrap成功応答以外のBrowser出力、生成HTML、URL、永続ブラウザ保存、ログ、fixtureに出さない。 | BR-015–016、NFR-001・003、DEC-ARCH-003 | explicit | session、HTTP、画面、設定、テスト資料で追跡できる。 |

## 資料間整合監査

- 要件→Feature設計→画面仕様→HTTP／operation資料を照合した。bootstrapでtokenを実行時メモリだけへ保持し、mutation／logoutにheaderで渡し、401・403・503時の画面遷移とtoken破棄・維持を定めている。画面、HTTP、operationの入出力・状態変更・エラーに矛盾はない。
- callbackはhashとciphertextを同時に保存し、bootstrapは復号後に同じhashで照合する。復号成功だけ、またはhash一致だけでtokenを返す経路は設計上存在しない。復号・hash照合の失敗を匿名成功へ変換せず503にするため、ciphertext改竄、保護鍵不整合、記録破損は認可の根拠にならない。
- `002`は既存`001`を変更・再適用せず、NULL可列の追加後にciphertextを持たない未失効sessionを失効する。migration中に失敗すればDDL、失効、version記録が同時にrollbackされ、成功後はlegacy recordが有効session判定を満たさない。bootstrapのcookie削除付き匿名応答、または管理操作の401により再ログインへ収束する契約と整合する。
- `csrf_token_ciphertext`をNULL可にするのは、既存schemaに対する追加migrationと、既に失効したlegacy rowを保持するためである。設計は新規sessionでciphertextを必須とし、NULLを認可・bootstrapに用いないため、nullableな物理列が新規資格の公開・運用意味を変えるものではない。
- DB access map、operation資料、test strategyを照合した。`002`の適用・再適用・rollback、NULL ciphertext legacy session失効、新規sessionのhash/ciphertext保存、bootstrap復号とhash不一致、HTTP 503、再ログインをunit／integration／HTTP／E2Eへ割り当てており、実装者へ安全性判断を委ねていない。
- 既存の保護サービスはPKCE verifier用にServer鍵によるauthenticated encryptionの`Seal`／`Open`を提供している。詳細設計はこの承認済み保護境界をCSRF ciphertextへ同じく使用し、暗号方式・鍵を新たな公開契約や利用者Decisionとして増やしていない。

## 人間Decision境界

DEC-FEAT-004のL3判断（migration時に既存sessionを失効し、一度だけ再ログインを求める）は承認済みである。ciphertextの平文保存、別CSRF方式、既存hashの置換など、未承認の互換方式を採用する記述はない。新たなL3/L4 Decisionは検出しなかった。

## 差し戻し事項

なし。

## 総合判定

**pass**

CSRF ciphertextの保存・復元・hash照合、`002_admin_session_csrf_ciphertext`のtransaction境界、legacy sessionの失効と再ログインへの収束は、承認済みDEC-FEAT-004と上流の認証・秘密情報境界に根拠を持ち、詳細設計・operation・DB／migration・HTTP・設定・検証資料間で矛盾なく実装可能である。

## 次ゲート

`implementation-readiness-review`。このpassはimplementation handoff可を意味しない。利用者による更新後の詳細設計の明示承認と、独立した実装準備性レビューのpassが必要である。
