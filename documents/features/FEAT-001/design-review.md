# FEAT-001 詳細設計レビュー

## 対象と独立性

`features/FEAT-001/{requirements.md,design.md,design/**,decisions/**}`を、設計者および過去のレビュー結論を根拠にせず、正規要件、承認済みDecision、上流設計へ反証的に照合した。今回の監査対象は、利用者が指定したBackendの`handler`／`usecase`／`repository`／`domain`への明示分割と、それに伴う既存の認証・DB・HTTP・Frontend契約の維持である。

## 正規ソース・Decision台帳

| 優先度 | ソース | 状態 | 適用範囲・確定内容 |
| --- | --- | --- | --- |
| 1 | `requirements/requirements.md` REQ-026–027 | confirmed | 事前登録済みGoogleメールアドレスの利用者だけが管理操作を行える。 |
| 1 | `requirements/business-rules.md` BR-001、BR-015–016 | confirmed | 管理者は許可メール1件の利用者だけであり、閲覧者・失敗表示へ管理情報、資格情報、秘密値、内部データを出さない。 |
| 1 | `requirements/nonfunctional-requirements.md` NFR-001、NFR-003 | confirmed | 認証情報、session、管理操作・管理情報を生成HTMLへ渡さない。 |
| 2 | `design/decisions/DEC-ARCH-001.md`、`DEC-ARCH-002.md` | 承認済み | 信頼済みWebアプリが認証・認可を所有し、生成HTMLは別originで隔離する。 |
| 2 | `design/decisions/DEC-ARCH-003.md` | 承認済み L3 | `cmd`が設定を読取り検証して注入し、`handler`／`usecase`／`repository`／`domain`の責務と禁止依存を明示する。repositoryの環境変数直接読取りは禁止される。 |
| 2 | `features/FEAT-001/decisions/DEC-FEAT-001.md`、`DEC-FEAT-002.md`、`DEC-FEAT-003.md` | 承認済み L3 | Server側session・CSRF・単一trusted origin、Backend独立Go moduleを採用する。 |
| 3 | `planning/{feature-map,traceability}.md`、`design/{architecture,cross-cutting-concerns,technology-constraints}.md` | 確定 | REQ-026–027はFEAT-001が所有し、4層の依存方向とFrontendのHTTP契約限定を採用する。 |

## 契約根拠の追跡

| 境界 | 設計上の契約 | 正規根拠 | 判定 | レビュー結果 |
| --- | --- | --- | --- | --- |
| 本人確認・認可 | Server側OIDCで検証済みのメールを、許可メール1件と完全一致で照合する。 | REQ-026–027、BR-001、DEC-ARCH-003 | explicit | callback、session、設定資料で追跡できる。 |
| session・CSRF | opaque session、同期CSRF token、状態変更時のtrusted Origin照合を必須にする。 | DEC-FEAT-001 | explicit | HTTP、operation、画面、DB設計で追跡できる。 |
| Gin HTTP境界 | Ginはrouting、middleware、HTTP request/response、cookie、redirect変換に限定する。 | DEC-ARCH-003 | explicit | `handler`だけがGinを扱い、`usecase`／`domain`へGin型を出さない設計になっている。 |
| 4層と物理配置 | `cmd`、`handler`、`usecase`、`repository`、`domain`を`backend/`直下で明示分離する。 | 利用者指示、DEC-ARCH-003、DEC-FEAT-003 | explicit | `design.md`、境界資料、test strategyで一致する。 |
| 依存・禁止依存 | `handler → usecase → domain`、`repository → usecase/domain`、`cmd`によるcompositionだけを許容する。 | DEC-ARCH-003 | explicit | 図、責務表、禁止依存、構造検証が整合する。`handler → repository`、`usecase → repository`実装、`repository → handler`、domainの外部依存を検出対象としている。 |
| 設定読取り・Secret | `cmd`が全Server設定を読取り・検証し、非秘密の認可依存をusecaseへ、Client Secret・保護鍵・DB接続情報をrepositoryへconstructor注入する。repositoryは環境変数・Secret storeを直接読まない。 | DEC-ARCH-003、DEC-FEAT-002 | explicit | `design.md`、`runtime-configuration.md`、`architecture-boundaries.md`、test strategyが同じ責務に揃っている。repositoryが注入済みのSecretを**使用**することと、環境から**読取る**ことを混同していない。 |
| PostgreSQL・Google外部I/O | `repository`がusecase portを実装し、外部の型・秘密値を内側へ返さない。 | DEC-ARCH-003 | explicit | port表、Google境界、DB access map、テスト戦略で追跡できる。 |
| HTTP・画面 | HTTP wire契約だけをFrontendが利用し、OAuth開始、callback、session初期化、logout、管理guardを画面状態へ対応付ける。 | DEC-ARCH-003、DEC-FEAT-001、DEC-FEAT-002 | explicit | HTTP、operation、画面仕様、test strategyにmethod/URI/status/cookie/画面遷移が追跡できる。 |
| 秘密情報非露出 | token、cookie値、許可メール、OAuth値、Secret、内部理由をBrowser、ログ、fixture、生成HTMLへ出さない。 | BR-015–016、NFR-001・003、DEC-ARCH-003 | explicit | session・HTTP・画面・外部境界・テスト資料で追跡できる。 |

## 資料間整合監査

- 要件→Feature設計→画面仕様→HTTP／operation資料を照合した。対象利用者、到達・退出、初期化・ログイン案内・認証済み・認証基盤障害、主要操作、API対応、アクセシビリティ、responsive、秘密情報非露出は追跡できる。
- `oauth_start`のform POST、Googleへの`303`、callback成功／失敗のcookie・遷移、session bootstrapとlogoutのHTTP契約は、operation・画面・テスト資料で追跡できる。PostgreSQL migration、保護記録の一回使用、Google timeout・retryなし、通常CIのtest doubleも対象資料間で矛盾しない。
- `design.md`、`architecture-boundaries.md`、`runtime-configuration.md`、test strategyを反証的に照合した。全資料が、環境変数・Secret storeの読取りと起動時検証を`cmd`だけの責務とし、repositoryはconstructor注入済みの設定・Secretだけを使い直接読取りを禁止している。前回の差し戻し原因だった設定責務の矛盾は解消済みである。
- `handler`はGin HTTP変換だけ、`usecase`はdomainと自身のportだけ、`repository`はusecase port実装と外部I/Oだけ、`domain`は業務規則だけを担当する。構造検証は4層の禁止依存と`cmd`以外の設定読取りを検出するため、実装者へ責務再設計を委ねていない。
- FrontendはBackend内部・PostgreSQL・Google・Secretへ依存せず、同一originのHTTP契約だけを利用する。生成HTMLの別origin結合検証をFEAT-005へ残す境界も、上流Decisionおよび要件と整合する。
- 公開契約、保存先・設定・運用責任、権限、外部連携について、承認済みDecisionまたは確定済み要件から導けない新しいL3/L4判断は検出しなかった。

## 差し戻し事項

なし。

## 総合判定

**pass**

`handler`／`usecase`／`repository`／`domain`の明示分割、4層の禁止依存、Gin・Google・PostgreSQL・OAuth・CSRF・session・Secret・Frontendの契約は、正規要件と承認済みDecisionに根拠を持ち、詳細設計資料間で一貫している。特に設定読取り・起動時検証を`cmd`に限定し、検証済み値をconstructor注入する契約は、repositoryの環境変数直接読取り禁止と矛盾しない。

## 次ゲート

`implementation-readiness-review`。pass後も、利用者による更新後の詳細設計の明示承認前にTask分割・implementation handoffへ進んではならない。
