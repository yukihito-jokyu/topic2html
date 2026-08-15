---
name: requirements-structuring
description: Convert a large, unstructured requirements document or collection of requirement notes into normalized system context, requirement IDs, business rules, non-functional requirements, constraints, and unresolved issues. Use before feature planning when requirements are raw, duplicated, inconsistent, or hard to trace. Do not decide product meaning when conflicting interpretations require a business decision.
---

# Requirements Structuring

## Goal

巨大・非構造な要件情報を、後続のFeature Planningと設計が機械的に参照できる形へ正規化する。

このSkillの目的は**要件を増やすことではなく、存在する要求を整理し、欠落・矛盾・未決定を可視化すること**。

## Preconditions

- 要件定義書、議事録、メモ、既存仕様のいずれかが存在する
- `.ai/workflow/decision-policy.md` が存在する
- `.ai/workflow/artifact-map.yaml` が存在する

## Inputs

優先順:

1. ユーザーが明示した最新要件
2. 正式な要件定義書
3. 承認済みDecision
4. 補助資料・議事録
5. 既存コードから推測できる現状仕様

推測したものはRequirementとして確定せず `Assumption` として扱う。

## Owned Outputs

- `requirements/system-context.md`
- `requirements/requirements.md`
- `requirements/business-rules.md`
- `requirements/nonfunctional-requirements.md`
- `requirements/constraints.md`
- `requirements/unresolved-issues.md`

必要に応じて:

- `requirements/decisions/DEC-*.md`

## ID Rules

- Functional Requirement: `REQ-###`
- Business Rule: `BR-###`
- Non-functional Requirement: `NFR-###`
- Constraint: `CON-###`
- Assumption: `ASM-###`
- Requirement Decision: `DEC-REQ-###`

既存IDがある場合は再採番しない。

## Autonomous Actions

人間確認なしで実施してよい。

- 要件候補の抽出
- 重複表現の統合候補作成
- Actor / 外部システム / System Boundaryの抽出
- Functional / Business Rule / NFR / Constraintの分類
- ID付与
- 同義語の正規化
- 矛盾候補の検出
- 要件間参照の整理
- 明示されていない実装詳細の除去またはConstraintへの移動

## Human Decision Points

`decision-policy.md` に従う。
特に以下は原則L4:

- Business Rule同士の矛盾
- 「何を実現するか」に関する複数の合理的解釈
- System Boundaryを変える判断
- 対象ユーザー・権限・責任主体を新たに決める判断

AIは勝手に一案へ確定しない。

## Procedure

### 1. Source Inventory

入力資料を列挙し、正規ソースの優先順位を明示する。

### 2. System Context抽出

最低限以下を整理する。

- Product / System Goal
- Actors
- External Systems
- In Scope
- Out of Scope
- Major Constraints

### 3. Atomic Requirement化

1つのRequirementに複数の独立した義務を混ぜない。
ただし、文言を細切れにして意味を失わせない。

各REQに最低限:

- ID
- Statement
- Source
- Rationale（分かる場合）
- Status: confirmed / assumed / unresolved

を持たせる。

### 4. Business Rule分離

「機能」ではなく複数操作に共通する制約・成立条件を `BR-*` として分離する。

### 5. NFR / Constraint分離

性能・セキュリティ・可用性・監査・互換性などはNFRへ、技術・組織・環境上の既定事項はConstraintへ置く。

### 6. Contradiction / Gap Analysis

次を検出する。

- 同じ対象に相反する条件
- Actorが不明
- 成功条件が不明
- 重要な異常系が未定義
- 保存・削除・権限・外部連携の責任主体が不明

推測で穴埋めできるL1/L2はAssumptionとして進める。
L3/L4はDecisionを発行する。

### 7. Unresolved Issues更新

未決定事項を重複なく一元化する。
解決済み項目は削除せず、解決先Decisionへのリンクを残してもよい。

## Guardrails

- UI、DB Schema、関数、ファイル名まで設計しない
- Featureへ分割しない。Feature分割は `feature-planning` の責務
- 「一般的にはこうだから」を理由にBusiness Ruleを追加しない
- 情報がないことと、禁止されていることを混同しない

## Exit Criteria

- System Contextが定義されている
- 主要なFunctional RequirementにIDがある
- Business Rule / NFR / Constraintが分離されている
- 重大な矛盾が未検出のまま残っていない
- 未決定事項は `unresolved-issues.md` またはDecisionとして明示されている
- 後続のFeature Planningが要件書全文を再解釈しなくても開始できる

## Next Skill

`feature-planning`
