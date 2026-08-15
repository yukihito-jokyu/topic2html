---
name: task-breakdown
description: Convert one completed feature design into implementation-ready logical tasks and an implementation handoff without prescribing technology-specific files, symbols, framework APIs, or code structure unless already fixed by approved design. Use after feature design and before implementation. Do not implement code and do not invent missing business or architectural decisions.
---

# Task Breakdown

## Goal

完成した1 Featureの設計を、Implementation領域がそのまま受け取れる**論理的な実装Task**へ分解する。

PlanningとImplementationの境界を守る。

このSkillが決めるのは主に **What / Responsibility / Acceptance / Dependency**。
Implementation Skillが決めるのは **How / File / Symbol / Library-specific implementation**。

## Preconditions

対象Featureについて `feature-design` のExit Criteriaを満たしている。

## Inputs

- `features/{feature_id}/requirements.md`
- `features/{feature_id}/design.md`
- `features/{feature_id}/design/**`（存在する場合）
- `features/{feature_id}/design-review.md`
- `features/{feature_id}/implementation-readiness-review.md`
- `.ai/workflow/state.yaml`
- 関連Decision
- `design/**`
- `planning/traceability.md`
- `.ai/workflow/implementation-handoff-schema.yaml`

## Owned Outputs

- `features/{feature_id}/tasks.md`
- `features/{feature_id}/implementation-handoff.yaml`

## ID Rules

- Task: `TASK-{feature-seq}-{task-seq}` または既存規則

プロジェクトでTask ID規約がある場合はそちらを優先する。

## Task Boundary

Taskは「作業者が成果を検証できる論理変更単位」とする。

良いTask:

- 目的が1つ
- Acceptanceが明確
- 依存関係が説明できる
- 複数技術スタックでも意味が維持される
- 実装者が設計そのものをやり直さなくてよい

悪いTask:

- `foo.goを編集する`
- `Repositoryを作る` だけで価値・責務が不明
- 1 TaskにFeature全体を詰め込む
- 1関数ごとに過剰分割する

## Autonomous Actions

- Feature Designから実装責務を抽出
- Task候補へ分割
- Task間Dependency整理
- Acceptance Criteria割当
- 並行実行可能性の識別
- Out of Scopeの引き継ぎ
- Implementation Handoff生成

## Human Decision Points

Taskの「粒度の好み」だけで人間を止めない。

以下を発見した場合のみ戻す。

- Feature Designに重大な仕様欠落がある
- Task化すると新しいBusiness Ruleを決める必要がある
- Initial Designに未決定のCross-feature contractがある

この場合、自分で補完せず `feature-design` または `initial-design` へ差し戻す。

## Procedure

### 1. Design Readiness Audit

Taskを生成する前に、対象 `design.md` のContract Completenessを監査する。

- `design-review.md` が存在し、設計者とは異なるsubagentによる総合判定 `pass`、正規ソースへの根拠トレーサビリティ、資料間整合の結果を持つことを確認する。存在しない、または `pass` 以外ならTaskを作成しない。
- `implementation-readiness-review.md` が存在し、設計者・design-review実行者と異なるfresh subagentによる総合判定 `pass`を持つことを確認する。存在しない、または `pass` 以外ならTaskを作成しない。
- workflow stateに、利用者が詳細設計を明示承認した記録があることを確認する。独立レビューの `pass`、沈黙、または閲覧だけを承認として扱わない。記録がなければTaskを作成せず、利用者レビューへ戻す。

- Featureの性質に対して必要な論理契約が `complete` または `not_applicable` か確認する。
- 「実装時に決める」「後で具体化」「方針未定」など、振る舞い・データ・契約の決定を実装Taskへ委ねる記述を検出する。
- 必要なフローチャートまたはシーケンス図が存在するか、不要理由が記録されているか確認する。
- SQLite／関係DBが採用済みなら、table、column、型、NULL、key、制約、index、migration version のDDL相当契約を確認する。
- JSON出力CLI／APIが採用済みなら、operation別の採用済み入力形式とsuccess／error field、型、必須性、enum、ネスト、繰返し・省略時の意味、代表例を確認する。CLIではexit codeとstdout／stderrの契約も確認する。
- 実装対象が新規の実行可能成果物、永続化driver、migration、または統合testを含む場合は、既存コードまたはInitial Designに、実装基盤が確定していることを確認する。少なくとも、採用言語／runtimeと依存管理、必要なdriver等の外部依存方針、責務を分ける配置規約、format・test・静的検査の共通手順を確認する。既存コードがこれらを一意に示す場合はその規約を根拠としてよい。
- Operation Documentation Coverage Gateが該当する場合、全operationの資料にI/O、状態変更、冪等性／競合、DB read/write、transaction／rollbackがあり、DB schema reference・relationship map・access mapと要求されたoperation別図への入口があることを確認する。operationごとに図が不要なら、理由が記録されていることを確認する。
- 画面を新設・変更する、または画面からの操作を要件に含むFeatureでは、画面設計仕様書に対象画面、状態、操作、API／operation対応、該当するアクセシビリティ・responsive・秘密情報非露出があり、`design.md`から参照可能であることを確認する。画面を持たない場合は不要理由を確認する。

いずれかに不備があればTaskを作成せず、欠落した契約・図・Decisionを具体的に示して差し戻す。Feature固有の契約不足は `feature-design` へ、実装基盤の未決定は `initial-design` へ差し戻す。Task内のAssumptionや実装判断で補ってはならない。

### 2. Design Obligations抽出

Feature Designから「実装後に成立していなければならないこと」を列挙する。

### 3. Logical Areasへ整理

必要に応じて以下のような技術非依存の領域へ整理する。

- presentation
- application
- domain
- persistence
- external integration
- security
- migration
- testing / verification

これは実装レイヤーを強制するものではない。

### 4. Task生成

各Taskに最低限:

- ID
- Title
- Purpose
- Related Requirements
- Logical Area
- Work Description
- Acceptance Criteria
- Depends On
- Out of Scope / Notes

を記載する。

### 5. 粒度検査

分割しすぎの兆候:

- Taskが単なるファイル編集指示
- Acceptance単独で確認できない
- ほぼ常に別Taskと同時変更が必要

大きすぎる兆候:

- 複数の独立したAcceptance結果を持つ
- 途中状態を安全に検証できる単位がある
- 複数担当が独立して進められる責務を含む

### 6. Dependency検査

依存理由を説明できない順序制約を削除する。

### 7. Implementation Handoff生成

`implementation-handoff-schema.yaml` を参考に、最低限以下を渡す。

- Feature identity
- Requirements / Business Rules
- Tasks
- Logical areas
- Acceptance Criteria
- Constraints
- Approved Decisions
- Interfaces / contracts
- Design readiness audit と契約成果物への参照
- Out of Scope

### 8. Technology Leakage Check

以下がTaskへ漏れていないか確認する。

- file path
- function/class名
- framework-specific API
- library choice
- code pattern

ただしInitial Designまたは既存の承認済み契約で固定済みなら参照してよい。

## Guardrails

- コードを書かない
- 実装Skillを作らない
- Tech Stackに合わせたfile構成へ変換しない
- 設計の欠落をTask内の推測で埋めない

## Exit Criteria

- 全TaskがFeature Design上の義務へtraceできる
- Acceptance Criteriaが各TaskまたはFeature全体に対応している
- Design Readiness Auditが成功し、実装者が仕様を再設計する未確定事項がない
- 新規実行可能成果物を作る場合、採用技術または既存技術規約、必要な依存方針、配置責務、共通検証手順がInitial Designまたは既存コードから参照できる
- 採用済みのSQLite／関係DBまたはJSON CLI／APIに必要な形式契約が、handoffから参照可能である
- Operation Documentation Coverage Gateが該当する場合、operation別資料・DBリファレンス・access map・図の監査結果がhandoffから参照可能である
- 独立した設計レビューが `pass` であり、handoffからレビュー成果物を参照できる
- 独立した実装準備レビューが `pass` であり、handoffから監査成果物を参照できる
- 利用者の詳細設計承認があり、handoffから承認記録を参照できる
- 不要なTechnology-specific detailがない
- Dependencyが説明可能
- `implementation-handoff.yaml` 単体でImplementation領域が計画内容を把握できる

## Next Skill

Implementation領域へ移る。

- 適切なImplementation Skillがない: `implementation-skill-builder`
- 既に存在する: 対応するImplementation Skill
