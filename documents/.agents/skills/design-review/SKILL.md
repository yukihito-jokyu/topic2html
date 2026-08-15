---
name: design-review
description: Independently review one feature's detailed design against authoritative requirements, approved decisions, and upstream design before task breakdown. Use as a fresh subagent review when a feature has public interfaces, persistence, configuration, migration, or other contracts whose rationale must be verified; do not redesign or implement the feature.
---

# Design Review

## Goal

対象Featureの詳細設計を、設計者とは独立した視点で正規ソースへ照合し、実装準備レビューへ進める品質かを判定する。実装handoffは、design-reviewとimplementation-readiness-reviewの`pass`に加えて、利用者による詳細設計の明示承認後にのみ可能である。

設計を補完しない。根拠のない仕様、未解決Decision、資料間の矛盾を検出し、Ownerへ差し戻す。

## Preconditions

- 対象Featureの詳細設計が存在する
- 正規の要件ソースと承認済みDecisionが参照可能である
- 設計を作成したAgentとは別のfreshなsubagentが実行する

## Inputs

優先順:

1. ユーザーが指定した原典（Issue、正式要件定義書）
2. 承認済みDecision
3. `requirements/**`
4. `planning/**` と `design/**`
5. `features/{feature_id}/requirements.md`、`design.md`、`design/**`

設計者の結論、既存監査の結論、期待する判定を入力として渡さない。必要な原典と成果物だけから再構成する。

## Owned Output

- `features/{feature_id}/design-review.md`

要件、設計、Task、handoffを変更しない。

## Procedure

### 1. 正規ソースとDecisionを台帳化する

各ソースについて識別子・節・確定／未決定・適用範囲を記録する。原典にない内容を後続資料だけで正当化しない。

### 2. 契約根拠を追跡する

公開I/O、永続化、設定・配置、Migration、エラー、権限、互換性、外部境界ごとに、設計上の契約、正規ソース、判定、レビュー結果を表で示す。

判定は `explicit`、`derived`、`assumption`、`decision_required`、`contradiction` のいずれかとする。`derived` は原典または承認済みDecisionから一意に導ける場合だけに使う。

### 3. 資料間整合を監査する

画面を新設・変更する、または画面からの操作を要件に含むFeatureでは、要件→Feature設計→画面設計仕様書→API／operation資料の対応を確認する。対象画面、利用者、到達・退出、主要領域・操作、loading／empty／error等の該当状態、API／operation対応、該当するアクセシビリティ・responsive・秘密情報非露出が追跡可能でなければ、Feature Designへ差し戻す。画面を持たない場合はnot_applicableの理由を確認する。

該当する場合、要件→Feature設計→ユースケース→operation資料→図→DB参照／SQL→handoffについて、field名、必須性、状態変更、エラー、transaction、検索条件、Migration versionを照合する。例が正規契約と異なる場合も不整合として記録する。複数operationで利用目的を達成するFeatureでは、ユースケース資料が目的・利用者・前提・操作列・各操作を使う理由・終了条件を平易に説明し、全operationを少なくとも一つのユースケースへ対応付けていることを確認する。CLI／APIの各操作列では、呼び出すoperation、採用済み入力形式、成功応答JSON、応答値から次操作の入力への受渡しを例で追跡でき、例が既存の公開契約だけを使用することを確認する。さらに、その操作列に操作漏れ・重複・責務混同・範囲外機能の混入・不要な公開操作・資料間矛盾がないかの監査結果を確認する。operation資料がある場合、利用者が理解するための用語・概要、バリデーション条件と結果、条件別の結果表現、SQL処理コメント、図の日本語ラベルとMermaid構文も確認する。フローチャートでは継続分岐が左、終了分岐が右の宣言順になり、transaction開始、rollback、書込み失敗を含む遷移がシーケンス図・バリデーション設計と一致することを確認する。複数SQL文のDB接続では、SQL連番とシーケンス図のSQLite Store矢印が一対一で対応すること、異なる失敗結果だけでなく同じerror codeでも失敗条件またはSQL番号が異なる結果が一つの分岐へまとめられず、`break` の早期終了として成功経路の外にあり、SQL番号・rollback有無とともに区別されていることも確認する。

### 4. 人間Decisionの境界を判定する

原典または承認済みDecisionがない限り、公開CLI/API field、互換性、既定動作、保存先・設定・運用責任、共有データモデル、権限、外部連携、プロダクト意味をAIが確定してはならない。

L3/L4なら必要なDecision、選択肢、影響範囲を差し戻し事項として示す。推奨案は提示してよいが、採用済みと記録しない。

### 5. 判定を記録する

`design-review.md` に対象・正規ソース一覧・根拠トレーサビリティ表・資料間整合・総合判定・差し戻し事項・次のゲートを記録する。総合判定は `pass`、`return_to_feature_design`、`human_decision_required` のいずれかとする。`pass` の次のゲートは「implementation-readiness-review」であり、handoff可ではない。

## Gate

次のいずれかがあれば `pass` にしてはならない。

- 公開または運用契約に、`explicit`または一意な`derived`ではない項目がある
- `assumption`が公開契約・データ意味・業務規則を確定している
- L3/L4 Decisionが未承認である
- 資料間に未解消の矛盾がある
- 実装者またはTaskへ仕様再設計が委ねられている

## Guardrails

- Featureの設計を直接修正しない
- 「一般的にはこうする」で新しい仕様を正当化しない
- 原典にない事項をRequirementへ昇格しない
- Task作成や実装をしない

## Exit Criteria

- 正規ソースと承認済みDecisionへの根拠追跡がある
- 人間DecisionとAIが導出可能な詳細が区別されている
- 資料間整合の結論と、利用者レビューへ進める可否が記録されている
- `pass`以外なら差し戻し先が明確である
