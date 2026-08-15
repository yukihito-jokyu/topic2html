---
name: initial-design
description: "Establish only the cross-cutting, high-change-cost design foundation needed before iterative feature development: system boundaries, architecture boundaries, domain boundaries, data ownership, technology constraints, interfaces, security/error/logging/testing principles. Use after feature planning for a greenfield or structurally changing project. Do not perform detailed design for every feature up front."
---

# Initial Design

## Goal

各FeatureをJust-in-Timeで詳細設計できるように、**後から変えると多数Featureへ影響する土台だけ**を先に決める。

Big Design Up Frontを避ける。
Feature固有の詳細は意図的に未決定のまま残す。

## Preconditions

- requirements-structuring完了
- feature-planning完了
- `.ai/workflow/decision-policy.md` が存在する

## Inputs

- `requirements/**`
- `planning/feature-map.md`
- `planning/feature-dependencies.md`
- 既存コード（存在する場合）
- `AGENTS.md` / 技術制約

## Owned Outputs

- `design/architecture.md`
- `design/domain-boundaries.md`
- `design/data-ownership.md`
- `design/cross-cutting-concerns.md`
- `design/technology-constraints.md`
- `design/decisions/DEC-ARCH-*.md`

## ID Rules

- Architectural / cross-cutting Decision: `DEC-ARCH-###`

## Initial Design Boundary

先に決める候補:

- System Boundary
- Runtime / deploy boundary
- Major component / layer responsibility
- Domain Boundary
- Data Ownership
- Authentication / Authorization principle
- External system boundary
- Shared interface policy
- Error / logging / observability principle
- Test strategy
- Deployment / configuration principle
- Technology constraints already committed

原則後回し:

- Feature固有の画面状態
- Feature固有のAPI形状
- 個別validation
- SQL詳細
- 関数・class・package設計
- 個別timeout値

## Autonomous Actions

- 要件・Feature群からcross-cutting concernを抽出
- Architecture候補を複数作る
- 候補のtrade-off比較
- 既存技術制約との適合性確認
- 後から変更コストが高い未決定事項を検出
- 設計しなくてもFeature開発を開始できる項目を意図的にDeferredにする

## Human Decision Points

L3/L4を人間へ上げる。
特に以下は原則Human Decision:

- Architecture boundaryの主要変更
- Domain ownership / data ownership
- Authentication / trust model
- 永続化・公開契約に大きく影響する選択
- 複数の有力技術選択肢のうち、長期制約を生む選択

AIは必ず候補比較と推奨を先に作る。

## Procedure

### 1. Cross-cutting Inventory

Feature Map全体を見て、複数Featureへ共通する設計問題だけを抽出する。

### 2. Design Drivers

Architecture判断を左右する要因を列挙する。

- Business Rules
- NFR
- Constraints
- External integrations
- Expected change axes

### 3. Minimal Architecture

必要最小限の責務境界を定義する。

図を使う場合も、名前だけの箱を増やさず責務・依存方向・禁止事項を記述する。

### 4. Domain / Data Ownership

主要な概念の所有境界を決める。
Feature固有のEntity詳細まで決めない。

### 5. Cross-cutting Policy

最低限、該当するものについて方針を記録する。

- errors
- logging / tracing
- security
- validation responsibility
- transaction principle
- testing
- configuration / secrets

### 6. Technology Constraints

「採用済み」「候補」「未決定」を区別する。
採用理由がDecisionに相当する場合は記録する。

### 7. Deferred Decisions

今決める必要のない設計を明示的にDeferredとして残す。
これにより「未記載 = 抜け」を防ぐ。

## Guardrails

- 全FeatureのSequence Diagramを作らない
- 全API/DB Schemaを先に固定しない
- 未来のFeatureを想像して過剰な抽象化をしない
- 技術選定をRequirementのように偽装しない
- Feature固有判断をcross-cutting policyへ昇格させない

## Exit Criteria

- Feature Designが依存できる責務境界が存在する
- Cross-featureな高コストDecisionが決定済みまたはPending Decisionとして明示されている
- 主要なData / Domain ownershipが不明瞭なままではない
- Deferred Decisionと設計漏れが区別されている
- Feature固有の詳細設計を開始できる

## Next Skill

`feature-design`
