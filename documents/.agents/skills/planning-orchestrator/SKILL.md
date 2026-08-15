---
name: planning-orchestrator
description: Orchestrate the project workflow from raw requirements through implementation-ready task handoff. Use when starting, resuming, or checking progress of requirements structuring, feature planning, initial design, feature design, or task breakdown. It owns workflow state and routes work to the appropriate planning skill. Do not use it to implement product code or to author another skill's design artifacts.
---

# Planning Orchestrator

## Goal

要件定義から `implementation-handoff.yaml` が完成するまでのPlanning領域を進行管理する。

このSkillは設計者ではなく**進行管理者**である。
要件・設計・タスクの内容を自分で代筆せず、各成果物のOwner Skillへ処理を委譲する。

## Scope

対象:

1. requirements-structuring
2. feature-planning
3. initial-design
4. feature-design
5. design-review
6. implementation-readiness-review
7. 人間による詳細設計レビュー
8. task-breakdown

対象外:

- 製品コードの実装
- 実装技術ごとの詳細手順
- 実装Skillの生成そのもの
- Skill定義の修正

対象外はそれぞれ `implementation-skill-builder`、生成されたImplementation Skill、`skill-maintainer` へ委譲する。

## Required Inputs

存在する場合は必ず読む。

- `.ai/workflow/artifact-map.yaml`
- `.ai/workflow/decision-policy.md`
- `.ai/workflow/state.yaml`
- `AGENTS.md`

## Bootstrap

`.ai/workflow/` が存在しない場合は、プロジェクトに以下を初期化してから進める。

- `.ai/workflow/artifact-map.yaml`
- `.ai/workflow/decision-policy.md`
- `.ai/workflow/state.yaml`
- `.ai/workflow/implementation-skills.yaml`
- `.ai/workflow/implementation-handoff-schema.yaml`

既存設定がある場合は上書きしない。
不足項目だけを補う。

## Artifact Ownership Rule

`artifact-map.yaml` を正規の配置ルールとして扱う。

このSkillが直接書き換えてよいProject Workflow Artifactは原則として:

- `.ai/workflow/state.yaml`

のみ。

以下を直接編集してはならない。

- `requirements/**`
- `planning/**`
- `design/**`
- `features/**`

内容変更が必要ならOwner Skillへルーティングする。

## State Machine

基本遷移:

```text
requirements
  -> feature_planning
  -> initial_design
  -> feature_loop
       -> feature_design
       -> design_review
       -> implementation_readiness_review
       -> human_design_review
       -> task_breakdown
       -> implementation_ready
```

複数Featureについて `feature_design -> design_review -> implementation_readiness_review -> human_design_review -> task_breakdown` を繰り返す。
全Featureを先に詳細設計してからTask化する必要はない。

## Procedure

### 1. 現在状態を復元する

`state.yaml` と実際の成果物を照合する。

状態ファイルと実ファイルが食い違う場合:

- 成果物の存在だけで `completed` と判断しない
- Owner SkillのExit Criteriaを満たしているか確認する
- 明白な状態記録漏れならstateを修正する
- 内容の不整合ならOwner Skillへ戻す

### 2. 未完了の最上流工程を特定する

以下の順で確認する。

1. requirements-structuring
2. feature-planning
3. initial-design
4. current feature の feature-design
5. current feature の design-review
6. current feature の implementation-readiness-review
7. current feature の人間による詳細設計レビュー
8. current feature の task-breakdown

上流が未完了なら下流へ進まない。

### 3. Pending Decisionを確認する

`pending_decisions` がある場合:

- L3/L4で本当に停止が必要か `decision-policy.md` と照合する
- 人間の回答が既に会話・成果物に存在するならOwner Skillへ反映を依頼する
- 未決定なら、決定要求を重複生成せず既存Decisionを提示する

### 4. 独立レビューを委譲する

Feature Designの後、Task Breakdownの前に `design-review` を必ず実行する。設計者と異なるfreshなsubagentへ委譲し、原典・承認済みDecision・成果物だけを渡す。設計者の結論、既存監査の結論、期待する判定を渡してはならない。

reviewが `pass` 以外ならimplementation-readiness-review、人間による詳細設計レビューおよびTask Breakdownへ進めない。`return_to_feature_design` はFeature Designへ、`human_decision_required` は人間Decisionへ差し戻す。

### 5. 実装準備レビューを委譲する

`design-review`が`pass`になった後、`implementation-readiness-review`を必ず実行する。設計者およびdesign-review実行者と異なるfresh subagentへ委譲し、実装者が仕様を再設計せずに着手できるかを監査させる。

`pass`以外なら、人間による詳細設計レビューとTask Breakdownへ進めない。`human_decision_required`では、先に`feature-design`へルーティングし、Decision Ownerが`features/{feature_id}/decisions/DEC-FEAT-###.md`へ承認待ちDecisionを記録してから利用者へ提示する。その他の判定はFeature DesignまたはInitial Designへ差し戻す。

### 6. 人間による詳細設計レビューを待つ

design-reviewとimplementation-readiness-reviewがともに `pass` になったら、詳細設計、補助資料、両レビュー結果を利用者が確認できる機会を明示的に設ける。この時点ではTaskを作成しない。

利用者から詳細設計を承認する明示的な回答を得るまで待つ。沈黙、閲覧しただけの反応、または独立レビューの `pass` を承認として扱ってはならない。修正指摘があれば `feature-design` へ戻し、必要なら再び独立レビューを行う。承認の事実と根拠をworkflow stateへ記録する。

### 7. Owner Skillへルーティングする

対象工程のSkillを読み、そのSkillのPreconditions / Inputs / Procedure / Exit Criteriaに従って処理する。

対象工程ごとに、そのOwner Skillを実行するsubagentを起動する。subagentの完了報告と必要な成果物を確認してから次工程へ進む。Orchestrator自身が「代わりにやっておく」ことは禁止。

独立レビューは、設計者・実装者・直前reviewerと異なるfresh subagentへ委譲する。review結果が`pass`でない限り、対応する下流工程へ進めない。

### 8. 完了後にstateを更新する

Owner SkillのExit Criteriaを満たしたことを確認してから更新する。

Feature Designを `completed`、Task Breakdownを `completed`、またはFeatureを `implementation_ready` にする前には、Owner Skillが記録したDesign Readiness Auditを確認する。永続化、Command / CLI / API、検索、更新、Migration、外部境界を含むFeatureでは、必要な論理契約と必要な図が complete または not_applicable であり、実装Taskへの仕様再設計の委譲がないことを確認する。SQLite／関係DBまたはJSON CLI／APIが採用済みなら、DDL相当およびoperation別wire schemaが complete で、handoffから参照可能であることも確認する。新規の実行可能成果物、永続化driver、migration、統合testを含むFeatureでは、Initial Designまたは既存コードに、採用言語／runtimeと依存管理、必要な外部依存方針、配置責務、format・test・静的検査の共通手順が確定し、handoffから参照可能であることも確認する。Operation Documentation Coverage Gateが該当する場合は、全operationのI/O・DB接続・図、DBリファレンス、DB access mapの監査結果が complete またはoperationごとの不要理由付きで、handoffから参照可能であることも確認する。さらに `design-review.md` と `implementation-readiness-review.md` が、設計者・相互のreview実行者と異なるsubagentによる `pass` であり、根拠トレーサビリティと実装準備監査結果をhandoffから参照できること、ならびに利用者の明示的な詳細設計承認がworkflow stateに記録されていることを確認する。

監査に不備がある場合、Orchestrator自身は設計成果物を修正せず、欠落を明示してOwner Skillへルーティングする。

例:

```yaml
features:
  FEAT-007:
    design: completed
    design_review: completed
    implementation_readiness_review: completed
    human_design_review: approved
    task_breakdown: completed
    implementation_handoff: ready
```

### 9. 次の行動を決定する

`implementation-handoff.yaml` が完成したFeatureは `implementation_ready_features` に追加する。

Planning領域としての処理はここで完了可能。
実装を依頼された場合のみImplementation領域へ渡す。

## Human Decision Points

Orchestrator自身は新しい設計Decisionを作らない。
各Owner Skillが発行したL3/L4 Decisionを進行上管理する。

以下は人間確認なしで自走する。

- 次に使うSkillの選択
- stateの同期
- 完了済み工程のスキップ
- featureごとの進捗更新

## Failure Handling

次の場合は停止して問題を明示する。

- artifact-mapにOwnerが存在しない成果物を複数Skillが更新している
- stateから現在工程を一意に復元できない
- OwnerではないOrchestratorが設計・要件・タスク成果物を直接編集しようとしている
- Skill間のPreconditionが循環している
- 同じDecisionが複数Ownerから矛盾した内容で要求されている

これらがSkill設計上の問題なら `skill-maintainer` の対象として記録する。

## Exit Criteria

少なくとも次のどちらかを満たす。

- 次に実行すべきOwner Skillが特定され、その処理へ移行した
- 1つ以上のFeatureが `implementation_ready` になり、stateへ反映された

## Next

Planning完了後:

- 実装Skillが不足している: `implementation-skill-builder`
- 実装Skillが存在する: Implementation領域へhandoff
- Skill運用上の不具合がある: `skill-maintainer`
