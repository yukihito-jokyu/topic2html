---
name: implementation-readiness-review
description: Independently audit one feature's detailed design from an implementer's perspective after design review and before user design approval. Verify that persistence, wire, operation, configuration, migration, test, and external-boundary contracts are complete enough to implement without redesigning them. Do not create tasks or implement code.
---

# Implementation Readiness Review

## Goal

詳細設計を、設計者・設計レビュー担当者とは異なる**実装者視点**で監査し、実装時に仕様を再設計せず着手できるか判定する。

このSkillは要件適合を主に確認する`design-review`を代替しない。`design-review`の`pass`後、利用者による詳細設計承認前に実行する独立ゲートである。

## Preconditions

- 対象Featureの詳細設計と補助資料が存在する。
- `features/{feature_id}/design-review.md`が、設計者と異なるfresh subagentによる`pass`である。
- 承認済みDecisionとInitial Designが参照可能である。
- 実行者は設計者およびdesign-review実行者と異なるfresh subagentである。

## Inputs

- `features/{feature_id}/requirements.md`
- `features/{feature_id}/design.md`、`design/**`、`decisions/**`
- `features/{feature_id}/design-review.md`
- `requirements/**`、`planning/**`、`design/**`
- `.ai/workflow/decision-policy.md`、`.ai/workflow/state.yaml`

## Owned Output

- `features/{feature_id}/implementation-readiness-review.md`

要件、設計、Decision、Task、handoffを変更しない。

## Procedure

### 1. 実装者モデルを作る

実装者が、追加の製品・業務・公開・セキュリティ判断をせずに何を作るかを復元する。不足は「もっともらしい実装」で補わない。

### 2. 契約完全性を監査する

該当するものを監査し、不要な場合は根拠を記録する。

- 関係DB: table、column、型、NULL、key、制約、index、migration version・順序・失敗時の扱い。
- API/HTTP/CLI: operationごとのmethod・endpointまたはtransport、入力、success/error、status/exit code、field型・必須性・省略意味、header/cookie、代表例。
- 画面: UIを新設・変更する、または画面からの操作を要件に含む場合、対象画面、表示・操作・状態遷移、API／operation対応、該当するアクセシビリティ・responsive・秘密情報非露出が画面設計仕様書で実装可能に定義されていること。画面を持たない場合は不要理由。
- operation: 状態変更、冪等性、競合、DB read/write、transaction/rollback、外部呼出し失敗時の補償・cleanup、図。
- 設定・秘密値: 所有者、Server限定範囲、起動時検証、環境別値、Browser・非信頼境界への非露出。
- 外部境界: callback/redirect、認証・認可、timeout/retry、test double、失敗時fail closed。
- 検証: unit/integration/e2eの責務、fixtureの秘密値非露出、共通format・lint・test手順。

### 3. 実装への委譲を検出する

「実装時に決める」「後で具体化する」などを検索し、挙動・永続化・wire・設定・運用・安全性の契約に関わるなら不合格とする。file path、symbol、局所的なコード構造だけは実装へ委譲してよい。

### 4. 判定を記録する

`implementation-readiness-review.md`へ、監査対象、実装者が依存する契約、項目別結果、欠落、差し戻し先、総合判定、次ゲートを記録する。

`human_decision_required`の場合は、必要なDecisionの論点、選択肢に必要な前提、推奨案の評価基準、承認後にFeature Designが補完する契約を明記する。このSkill自身はDecisionを作成しない。DecisionのOwnerである`feature-design`が`features/{feature_id}/decisions/DEC-FEAT-###.md`を作成してから、利用者へ提示する。

総合判定は`pass`、`return_to_feature_design`、`return_to_initial_design`、`human_decision_required`のいずれかとする。

## Gate

次のいずれかがあれば`pass`にしてはならない。

- 採用済み関係DBのDDL相当契約またはmigration契約がない。
- 採用済みAPI/HTTP/CLIのoperation別wire contractがない。
- 状態変更・外部呼出し・永続化のtransaction/rollback/cleanupが実装者判断に委ねられている。
- 設定値、origin、callback、秘密値の所有・検証・運用責任が不明である。
- 実装前に人間が決めるべきL3/L4事項が未解決である。
- 画面を新設・変更する、または画面からの操作を要件に含むFeatureなのに、画面設計仕様書または不要理由がない。

## Next Gate

- `pass`: 利用者による詳細設計レビュー。
- `return_to_feature_design`: Feature Designへ差し戻す。
- `return_to_initial_design`: Initial Designへ差し戻す。
- `human_decision_required`: `feature-design`へ戻し、承認待ちDecisionを`features/{feature_id}/decisions/DEC-FEAT-###.md`に記録してから利用者Decisionへ進む。
