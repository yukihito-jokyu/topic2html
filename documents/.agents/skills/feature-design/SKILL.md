---
name: feature-design
description: Perform just-in-time detailed design for exactly one feature using normalized requirements and the existing initial design. Use when a feature is selected for implementation preparation. Decide which analysis techniques are actually needed (flows, state transitions, sequence, domain rules, interface/data/error/test design) and produce implementation-neutral feature design. Do not redesign cross-cutting architecture directly or split work into code-level tasks.
---

# Feature Design

## Goal

選択された**1 Featureだけ**を、Task Breakdownが機械的に行える粒度まで詳細化する。

このSkillは設計手法を一律に適用しない。
Featureの性質を分析して、必要な設計だけを行う。

## Preconditions

- `feature-planning` 完了
- `initial-design` 完了
- 対象 `feature_id` が決まっている

## Inputs

必須:

- `requirements/**`
- `planning/feature-map.md`
- `planning/traceability.md`
- `design/**`
- 対象Featureに関連する既存Decision
- 既存コード（存在する場合、現行契約の確認用）

## Owned Outputs

対象 `{feature_id}` について:

- `features/{feature_id}/requirements.md`
- `features/{feature_id}/design.md`
- `features/{feature_id}/design/**`（操作別・DB参照など、必要な補助設計資料）
- `features/{feature_id}/decisions/DEC-FEAT-*.md`
- `features/{feature_id}/design-change-requests/CR-*.md`

## ID Rules

- Feature-local Decision: `DEC-FEAT-###`
- Design Change Request: `CR-###`

## Analysis Selection

最初に、以下の分析が必要か判定する。

- Use case / scenario
- Main / alternate / error flow
- State transition
- Sequence / interaction
- UI state
- Screen / UI specification
- Application responsibility
- Domain rule
- Interface contract
- Data change
- Persistence behavior
- Error model
- Security/privacy concern
- Concurrency / idempotency / transaction
- Test strategy
- Flowchart
- Sequence / interaction diagram

不要な分析を儀式として生成しない。

例:

- 単純なread-only lookup: 状態遷移図は通常不要
- Wizard / workflow: 状態遷移が重要
- 複数component連携: sequenceが有用
- 複雑なBusiness Rule: domain ruleを明示
- 分岐、再試行、停止条件、状態更新または例外復帰が実装判断を左右する: flowchartが有用

### Diagram Policy

図は `features/{feature_id}/design.md` に Mermaid で記載する。

- フローチャート: 分岐、再試行、停止条件、状態更新、例外復帰などの処理ロジックがある場合に作成する。
- シーケンス図: 複数の責務境界、外部システム、永続化境界をまたぐ呼出し順序、データ受渡し、失敗時の戻り先が実装判断を左右する場合に作成する。
- 上記が不要な場合は、不要と判断した理由を記録する。単純な CRUD や単一責務の変換に図を儀式として追加しない。

利用者がコマンド／operation単位のフローチャートまたはシーケンス図を要求した場合、その要求は当該Featureの設計完了条件とする。共通図だけで個別operationの分岐、永続化境界、失敗時の戻り先を代替してはならない。各operationに図が不要な場合は、operation名と不要理由を明記する。

## Autonomous Actions

- 対象FeatureのRequirement再構成
- 正常・代替・異常シナリオ列挙
- 必要な設計手法の選択
- Feature内の責務分解
- Feature内のL1/L2 Decision
- Edge Case抽出
- Acceptance Criteria具体化
- Test観点整理
- Existing Initial Designへの適合性確認

## Human Decision Points

以下の場合は人間判断またはOwner Skillへの差し戻しが必要。

### Feature内だがProduct意味を変える

L4として人間へ。

### Initial Design変更が必要

自分で `design/**` を変更しない。

`design-change-requests/CR-*.md` を作成し、以下を記録する。

- 現在の設計
- なぜFeature実現を阻害するか
- 必要な変更
- 影響Feature
- 推奨

その後 `initial-design` へ戻す。

### 複数Featureの契約変更

L3として扱う。

### L3/L4のFeature固有Decision

Feature内のL3/L4判断は、承認待ち・承認済みを問わず
`features/{feature_id}/decisions/DEC-FEAT-###.md` に記録する。
`design/**` の補助資料として置かない。Decision本文には状態、選択肢、推奨、承認後の影響を記載する。
Decision本文へoperation別I/O、状態変更、図、検証設計を混在させない。それらは承認後に
`design.md`から参照する操作別・共通契約の補助資料へ分離する。

## Procedure

### 1. Feature Scope固定

対象FeatureのGoal、Included Requirements、Out of Scopeを確定する。

### 2. Behavioral Design

利用者またはTriggerから観測可能な結果までを設計する。

最低限:

- Trigger
- Preconditions
- Main flow
- Alternate/error flows
- Postconditions

### 3. Required Analysisを選択

選択した分析と、不要と判断した分析を簡潔に記録する。
「なぜ図を作らなかったか」が後から分かる状態にする。

### 4. Responsibility Design

実装技術名ではなく論理責務として整理する。

例:

- input validation
- orchestration/application operation
- domain decision
- external capability
- persistence
- presentation state

### 5. Contract / Data / Error

Featureが境界を跨ぐ場合に必要な契約を定義する。

ただし、Initial Designで技術が固定されていない限り、framework固有の型やfile pathへ落とさない。

### 5a. Contract Completeness Gate

Featureが次の性質を持つ場合、該当する論理契約を `design.md` で具体化する。実装者またはTaskに仕様の決定を委ねてはならない。

| Featureの性質 | 必須の論理契約 |
|---|---|
| 永続データまたは履歴を持つ | 概念ごとの属性、必須性、安定識別子、関連・多重度、現行／履歴の扱い、削除・保持方針 |
| Command / CLI / API 境界を持つ | 操作ごとの入力、正常出力、構造化エラー、validation、状態変更、互換性方針 |
| 検索・一覧・取得を持つ | 検索条件、Scope・filter、照合規則、順序、件数制限／継続取得、空結果、not-found の区別 |
| 更新・重複の可能性を持つ | 同一性・重複の規則、冪等性、競合時の扱い、原子性 |
| Schema / Store / Index を導入または変更する | 初期化、配置・設定の論理契約、版管理、適用順、失敗・中断時の扱い、互換性・復旧方針 |
| テストでデータ状態を検証する | 代表 Fixture、操作、期待する状態・出力・エラー |

採用済みの技術・transport がある場合は、上記に加えて次の**形式契約**を `design.md` に記載する。これは file path、symbol、framework API、library-specific の実装指示ではない。承認済み技術上での相互運用可能なデータ／wire 契約である。

| 固定済みの前提 | 必須の形式契約 |
|---|---|
| SQLite または他の関係DB | 論理テーブルごとの table 名、column 名、SQLite affinity または論理型、NULL可否、primary key、foreign key、unique 制約、check 制約、検索に必要な index、各 migration の version・依存・前後条件。 |
| JSON出力 CLI / JSON API | operation ごとの入力形式（名前付きoption、request JSONなど採用済みtransport）と、success／error response の JSON Schema または同等のfield表。すべての入力・出力field名、型、必須性、enum、ネスト、繰返し・省略時の意味、代表例を記載し、CLIならexit codeとstdout／stderrの割当も記載する。 |

技術または transport が未決定なら、形式契約は `not_applicable` として理由と未決定のDecisionを記録する。技術・transport が採用済みなのに「実装時にテーブル／field 名を決める」とすることは未完了である。

### 5b. Operation Documentation Coverage Gate

JSON CLI／APIと永続Storeを併せ持つFeature、または利用者がコマンド単位で追跡可能な設計資料を要求したFeatureでは、全operationを一覧化し、実装者がoperation単位で参照できる設計資料を作成する。資料の読みやすさは設計品質の一部であり、巨大な表へ情報を重複させない。

- `design.md` は入口として、Featureの全体方針、共通契約、資料一覧、各補助資料へのリンクを持つ。
- 必要に応じて `features/{feature_id}/design/` 配下へ補助資料を配置する。標準構成は command catalog、`commands/{operation}.md`、database schema reference、database relationship map、database access map とする。単一 `design.md` で可読性を損なわず満たせる場合は分割不要だが、その理由を記録する。
- command catalog は提供operation、分類、各operation資料へのリンク、共通JSON envelope／error／exit codeへの参照を持つ。
- 複数operationを組み合わせて利用目的を達成するFeatureでは、ユースケース別の資料を `design/use-cases/{use-case}.md` に配置する。各ユースケース資料には、何の機能か、利用者、前提、目的、根拠、目的を達成する操作列、各操作をその順で使う理由、成功時と操作列を終える条件を平易に記載する。CLI／APIでは、各順序について「呼び出すoperation」「採用済み入力形式の例」「成功応答JSONの例」「応答のどの値を次操作のどの入力へ渡すか」をそのユースケース資料で示す。ユースケース横断の操作対応・過不足・矛盾監査も、いずれかのユースケース資料または利用者が求めた索引資料へ記録する。既存の `design/use-cases.md` を索引として残すか、ユースケース別資料へ完全移行して削除するかは利用者の指示に従う。例は承認済みの公開契約だけを使い、未根拠の引数・field・設定・保存先を追加しない。全operationを少なくとも一つのユースケースへ対応付け、要件・Decision・operation資料と照合して、操作漏れ、重複、責務混同、範囲外機能の混入、資料間矛盾、不要な公開操作がないかを監査結果として記録する。操作列が不要なFeatureでは、不要理由を記録する。
- JSON出力CLIの各operation資料は、原則としてタイトル直後に `## 用語`（初出の業務語・fieldの意味を平易に説明）と `## コマンド概要`（利用者視点の目的・処理結果）を置き、続けて `## I / O`、`## バリデーション設計`、`## フローチャート`、`## シーケンス図`、`## DB接続` の順に記載する。I / Oは採用済み入力形式の例とsuccess responseの**インデント付き複数行の**JSON code blockを示し、成功時の出力意味、終了コード・stdout／stderrを本文で明記する。入力の型・空値・enum・重複・参照先の有無・競合と、それらに対応する error response はバリデーション設計にだけ記載し、I / Oへ重複させない。結果条件・エラー条件は、条件ごとの箇条書きまたは表で分離し、長い条件連結文にしない。バリデーション設計には入力の型、空値、enum、重複、not-found、conflict、storage failureのうち該当する条件と結果を記載する。DB接続はread/write、transaction・rollbackを本文で説明し、承認済みSQL構文で具体化できる場合はSQL code blockも示す。SQL blockには処理目的を説明するコメントを付ける。
- database schema reference はtableごとの目的、column、key、constraint、index、他tableとの参照関係を説明する。DDLのみを貼り付けて完了扱いにしない。
- database access map はoperationごとにread／writeするtableと主要column、派生Index、transaction有無を逆引きできるようにする。
- operation固有の分岐、状態変更、例外復帰がある場合はflowchartを作る。CLI／API、Ledger、Store、Indexなどの責務境界をまたぎ、順序または失敗時の戻り先が実装判断を左右する場合はsequenceを作る。利用者がoperation別の図を要求した場合は、各operationについて作成するか不要理由を記録する。
- フローチャートでは、判断ノードから処理が継続する分岐を先に宣言して左側へ、終了する分岐（error response、早期終了、rollback後の終了）を後に宣言して右側へ配置する。分岐順の変更だけで完了とせず、正常・異常の遷移、transaction開始、rollback、書込み失敗がシーケンス図・バリデーション設計と整合することを確認する。
- 図の利用者向けラベルは日本語で記載する（公開済みのoperation名、JSON field名、error codeなど正確さを要するリテラルは除く）。Mermaidの記法にはMarkdown装飾を混在させず、構文上安全なラベルを使う。DB接続に複数のSQL文がある場合は各文へ連番を付け、シーケンス図でも各SQL文について同じ番号を付けたSQLite Storeへの矢印を個別に記載する。シーケンス図の失敗分岐は、異なるerror codeだけでなく、同じerror codeでも失敗した条件またはSQL番号が異なるなら別々に記載する。失敗分岐は `break` を使う早期終了として成功経路の外に置き、各分岐で、返すerror code、対応するSQL番号、rollbackの有無を示す。

このGateはfile path、symbol、framework API、driver固有APIを決めるものではない。資料の論理構造と、承認済みSQLite／JSON境界の相互運用契約を定義するものである。

### 5c. Screen Design Coverage Gate

Featureが信頼済みアプリケーションの画面を新設・変更する、または画面からの操作を要件に含む場合は、画面設計仕様書を `features/{feature_id}/design/` 配下に作成し、`design.md`から参照する。画面を持たないFeature、または既存画面の観測可能な挙動を一切変更しないFeatureでは `not_applicable` とし、理由を記録する。

画面設計仕様書には、少なくとも次を記載する。

- 対象画面・利用者・到達条件・退出先
- 画面ごとの目的、主要領域、表示要素、操作、操作可能条件
- loading、empty、success、validation error、authorization error、external/service failureなど該当する表示状態と遷移
- 画面操作とAPI／operationの対応、送る入力、受け取る結果、画面上の反映
- 要件に必要なアクセシビリティ、キーボード操作、responsive表示、秘密情報を表示しない境界

画面コンポーネントのfile path・symbol・framework APIや、要件に根拠のないpixel単位の見た目を固定してはならない。視覚デザインの別仕様が正規ソースとしてある場合だけ、それを参照する。

「後で具体化する」「実装時に決める」記述が、上記の振る舞い・データ・契約に関わる場合は未完了とする。L1/L2で一意に導ける内容は設計者が決め、L3/L4はDecisionまたはChange Requestとして解消する。

### 6. Acceptance / Test Design

Task BreakdownがAcceptance Criteriaを再解釈しなくてよいレベルまで具体化する。

### 7. Consistency Check

以下を確認する。

- Requirementとのtrace
- Business Ruleとの矛盾なし
- Initial Designとの矛盾なし
- Out of Scopeの侵食なし
- 新しいBusiness Ruleを暗黙追加していない
- 該当するContract Completeness Gateがすべて complete または not_applicable である
- SQLite／関係DBまたはJSON CLI／APIが採用済みの場合、対応する形式契約が complete である
- Operation Documentation Coverage Gateが該当する場合、operation資料、DBリファレンス、DB access map、要求されたoperation別図、および該当時のユースケース資料・操作列監査が complete または不要理由付きである
- Screen Design Coverage Gateが該当する場合、画面設計仕様書があり、画面状態・操作・API／operation対応・該当する非機能要件が追跡可能である。またはnot_applicableの理由が記録されている
- 実装者またはTaskへ仕様再設計を委ねる記述がない

## design.md Minimum Sections

- Feature Summary
- Scope / Out of Scope
- Related Requirements / Business Rules
- Behavioral Scenarios
- Selected Design Analyses
- Responsibilities
- State / Interaction（必要な場合）
- Interfaces / Data
- Contract Completeness（該当する場合）
- Physical / Wire Schema（SQLite／関係DBまたはJSON CLI／APIが採用済みの場合）
- Operation Documentation（Operation Documentation Coverage Gateが該当する場合）
- Screen Design（Screen Design Coverage Gateが該当する場合）
- Error / Edge Cases
- Security / NFR considerations
- Acceptance / Test Design
- Assumptions
- Decisions
- Open Issues

## Guardrails

- 1 Featureを超えて詳細設計しない
- `internal/...` のようなfile pathを原則決めない
- function/class/package単位へTask化しない
- Initial Designを直接書き換えない
- 既存コードが偶然そうなっているだけの実装詳細をRequirementへ昇格しない

## Exit Criteria

- Featureの振る舞いが正常・主要異常系まで説明できる
- 必要な責務・境界・データ・エラーが明示されている
- 該当する論理契約が、実装者が仕様を再設計せずに実装できる粒度で complete または not_applicable と記録されている
- 採用済みのSQLite／関係DBまたはJSON CLI／APIについて、DDL相当およびoperation別wire schemaが complete と記録されている
- 必要なフローチャート・シーケンス図が `design.md` またはそこから参照する補助設計資料にあり、不要な場合は理由が記録されている
- Operation Documentation Coverage Gateが該当する場合、operation別I/O・DB接続・図の網羅状況、DBリファレンス／access mapへの入口、および該当時のユースケースによる全operationの利用目的・操作列・監査結果が記録されている
- Screen Design Coverage Gateが該当する場合、画面設計仕様書がdesign.mdから参照可能であり、画面状態・操作・API／operation対応・該当する非機能要件を実装者が再設計せずに実装できる
- Acceptance Criteriaが実装後に観測・検証可能
- L3/L4未決定がTask Breakdownを阻害する形で残っていない
- Task Breakdownが「仕様を再設計」せず分解可能

## Next Skill

`task-breakdown`
