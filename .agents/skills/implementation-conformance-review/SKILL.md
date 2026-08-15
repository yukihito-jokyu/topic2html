---
name: implementation-conformance-review
description: topic2htmlの実装後に、issueまたはTask、要件、承認済み設計・Decision、実装diff、テスト結果を独立に照合し、仕様逸脱・責務境界違反・未検証の受け入れ基準を検出する。コードを変更せず、実装完了前の独立監査が必要なときに使用する。
---

# 実装整合レビュー

## Goal

実装担当者とは異なるfresh subagentが、実装がissueまたはTaskと承認済みの要件・設計仕様書・Decisionから別方向へ進んでいないかを反証的に監査する。

## Required Inputs

- issue本文・受け入れ基準、または対象Task
- `documents/features/{feature_id}/implementation-handoff.yaml`
- 対象Featureのrequirements、design、decisions、design-review、implementation-readiness-review
- 実装diff、変更ファイル、テスト・lint・buildの実行結果
- backendを変更した場合は、`task backend:coverage`またはTask規約で明記された同等スクリプトと、その実行結果
- 親`AGENTS.md`、現行コード、関連する`impl-*` Skill

issueが存在しない場合はTaskを一次根拠にし、その旨を報告する。外部trackerの取得権限がない場合、与えられたissue本文以上を推測しない。

## Procedure

1. issueまたはTaskの受け入れ基準を、要件・設計・Decision・handoffへ対応付ける。
2. diffと変更後コードを読み、対象外機能、未承認dependency、公開contract、DB migration、認可・秘密情報・隔離境界の逸脱を探す。
3. 受け入れ基準ごとに、実装箇所と適切なunit / integration / HTTP / E2E検証の証拠を確認する。
4. 正常系だけでなく、設計上の失敗経路、rollback、未認可、外部障害、秘密情報非露出を確認する。
5. backendを変更した場合は、実装者と同じカバレッジ検証入口が存在することを確認し、freshなreviewerとして実行する。table-driven testが必要な複数caseの検証を単発testへ退行させていないかも確認する。
6. `pass`、`changes_required`、`return_to_planning`、`human_decision_required`のいずれかを根拠とともに報告する。コード・設計成果物を編集しない。

## Gate

次のいずれかがあれば`pass`にしない。

- issue / Taskの受け入れ基準に、実装または検証の対応付けがない。
- 承認済みの要件・設計・Decisionと矛盾する挙動、公開contract、永続化、認可、秘密情報の扱いがある。
- 実装が対象外Featureの業務契約や隔離境界を先取りしている。
- 実行済みとされた検証の証拠がない、または失敗を隠している。
- backend変更時に共通カバレッジ検証入口が存在しない、reviewerが実行できない、または100%未満・非0終了である。

## Independence

対象Featureのfeature-design担当、design-review担当、implementation-readiness-review担当、実装担当、実装前reviewerの全てと異なるfresh subagentが実行する。親agentの実装方針を正しいと仮定せず、差分と一次資料から反証する。

## Completion Criteria

- 判定と根拠、受け入れ基準への対応、検証済み範囲、残存リスクを報告した。
- `pass`以外では実装完了を宣言しない。
