---
name: implementation-preflight-review
description: topic2htmlで実装を開始する前に、issueまたはTask、implementation handoff、承認済み要件・設計・Decision、現行コードを独立に照合し、実装の開始条件とスコープ逸脱を監査する。コードを書かず、製品仕様を決めず、実装前の独立ゲートが必要なときに使用する。
---

# 実装開始前レビュー

## Goal

実装担当者とは別の視点で、実装要求が承認済みのissueまたはTask、handoff、要件、設計仕様書、Decision、現行コードと整合し、仕様を再設計せず安全に着手できるかを判定する。

## Required Inputs

- issue本文・受け入れ基準、または対象Task
- `documents/features/{feature_id}/implementation-handoff.yaml`
- 対象Featureのrequirements、design、decisions、review記録
- `documents/.ai/workflow/state.yaml`と親`AGENTS.md`
- 既存コード、類似実装、実行可能な検証設定

issueが存在しない場合はTaskを一次根拠にし、その旨を報告する。外部trackerの取得権限がない場合、与えられたissue本文以上を推測しない。

## Procedure

1. handoffの`implementation_ready`とworkflow stateの整合を確認する。
2. issueまたはTaskの目的・受け入れ基準を、要件、承認済み設計、Decision、handoffへ対応付ける。
3. 現行コードと設定を読み、依存Task、既存contract、対象外範囲、利用すべき`impl-*` Skillを確認する。
4. 未承認の業務・公開・security・architecture判断、未解決Decision、仕様矛盾、実装順序の欠落を探す。
5. `pass`、`return_to_planning`、`human_decision_required`のいずれかを根拠とともに報告する。成果物を編集しない。

## Gate

次のいずれかがあれば`pass`にしない。

- handoffとworkflow stateが不整合、またはFeatureがimplementation-readyでない。
- issue / Taskの受け入れ基準を根拠資料へ対応付けられない。
- 実装で新しい業務、公開contract、security posture、architectureを決める必要がある。
- 必須の依存Task、設定、migration、test strategyが欠ける。

## Independence

対象Featureのfeature-design担当、design-review担当、implementation-readiness-review担当、実装担当の全てと異なるfresh subagentが実行する。実装方針の正当化ではなく、仕様逸脱と開始不能な前提を反証的に探す。

## Completion Criteria

- 判定と根拠、確認した資料、ブロッカーまたは残存リスクを報告した。
- `pass`以外では実装開始を許可しない。
