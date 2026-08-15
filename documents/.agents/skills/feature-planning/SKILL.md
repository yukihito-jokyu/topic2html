---
name: feature-planning
description: Transform normalized requirements into a feature map, feature boundaries, dependencies, backlog candidates, and requirement-to-feature traceability. Use after requirements structuring and before initial design or per-feature detailed design. It may propose MVP and ordering, but must not silently decide product priority or split features based only on code-layer boundaries.
---

# Feature Planning

## Goal

正規化された要件群を、価値・振る舞いのまとまりとしてFeatureへ分割し、実装順序を検討できるBacklogへ変換する。

Featureはコードレイヤーではなく、**利用者またはシステムにとって意味のある能力**として定義する。

## Preconditions

以下が `requirements-structuring` のExit Criteriaを満たしている。

- `requirements/system-context.md`
- `requirements/requirements.md`
- `requirements/business-rules.md`

## Inputs

必須:

- `requirements/**`
- `.ai/workflow/decision-policy.md`
- `.ai/workflow/artifact-map.yaml`

## Owned Outputs

- `planning/feature-map.md`
- `planning/feature-dependencies.md`
- `planning/backlog.md`
- `planning/traceability.md`

必要に応じて:

- `planning/decisions/DEC-PLAN-*.md`

## ID Rules

- Feature: `FEAT-###`
- Planning Decision: `DEC-PLAN-###`

## Feature Boundary Principles

良いFeatureは概ね以下を満たす。

- 独立した利用価値またはシステム能力として説明できる
- AcceptanceをFeature単位で定義できる
- Requirementとの対応関係を追える
- 「frontend」「repository」のようなコード層そのものではない
- 巨大すぎる場合はユーザーシナリオまたは状態変化で分割できる

## Autonomous Actions

- RequirementをFeature候補へグルーピング
- Feature名と目的の提案
- Requirement -> Feature traceability作成
- Feature間の前提・依存関係抽出
- 大きすぎるFeatureの分割案提示
- 重複Feature候補の統合
- 技術的な先行調査が必要なSpike候補の識別
- 実装順序の推奨案作成

## Human Decision Points

原則人間判断:

- MVP / Release Scope
- ビジネス価値による優先順位
- Feature境界を変えることでProduct UXの意味が変わる場合
- どの利用シナリオを先に提供するかが事業判断になる場合

純粋な依存関係から順序が一意に決まる場合は人間確認不要。

## Procedure

### 1. Requirement Clusterを作る

Actor、目的、対象データ、状態変化、業務ルールを手掛かりにRequirementをまとめる。

### 2. Feature Candidateを作る

各Featureに最低限:

- Feature ID
- Name
- Goal
- Primary Actor / Trigger
- Covered Requirements
- Related Business Rules
- High-level Acceptance
- Dependencies
- Status

を持たせる。

### 3. Boundary Review

次のアンチパターンを検出する。

- 1 Feature = 1 Layer
- Featureが「DBを作る」「APIを作る」のような手段だけになっている
- 1 Featureに複数の無関係な利用価値が混ざっている
- RequirementがどのFeatureにも属さない
- 1つのRequirementが不必要に多数Featureへ重複している

### 4. Dependency Graph

依存関係を以下のように区別する。

- behavioral prerequisite
- data prerequisite
- platform prerequisite
- optional sequencing

単なる「この順が作りやすい」はHard Dependencyにしない。

### 5. Backlog化

Backlogには少なくとも:

- Feature ID
- Name
- Status
- Priority（確定またはproposed）
- Dependency
- Planning Notes

を記録する。

### 6. Traceability更新

Requirement / Business RuleとFeatureの対応を明示する。

## Guardrails

- DB Table、class、package単位へ分解しない
- タスク化しない
- Featureごとの詳細フローを作り込まない
- 優先順位を「一般的だから」で勝手に確定しない

## Exit Criteria

- 主要Requirementが少なくとも1つのFeatureへtraceされている
- Featureが価値・振る舞い単位で定義されている
- Hard Dependencyと単なる推奨順序が区別されている
- 次にInitial Designで考慮すべきシステム横断事項が見える
- 最初に詳細設計するFeature候補を選べる状態になっている

## Next Skill

`initial-design`
