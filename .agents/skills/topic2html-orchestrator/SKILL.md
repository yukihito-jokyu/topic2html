---
name: topic2html-orchestrator
description: topic2htmlの要件整理からimplementation-ready handoffまでのPlanning workflowを開始・再開・進捗確認する。要件、Feature分割、Initial Design、Feature Design、独立レビュー、利用者承認、Task分割を依頼されたときに使用し、プロダクトコードや実装Skill自体は変更しない。
---

# topic2html Orchestrator

## Goal

`documents/`にある正本workflowを使い、要件から実装完了前の独立監査までを正しいSkillへルーティングする。各工程はSkillごとにsubagentへ委譲し、自身で要件・設計・Task・コードを代筆しない。workflow stateだけを進行管理する。

## Use When

- Planning workflowの開始、再開、現在工程の確認を依頼された。
- 要件整理、Feature分割、Initial Design、Feature Design、Task分割を依頼された。
- Featureの詳細設計レビュー、実装準備性レビュー、利用者承認、handoff、実装前後の独立レビューの進行を依頼された。

## Do Not Use When

- 実装コードやテストだけを変更する。
- Go / React / PostgreSQLの実装手順を作る（対応する`impl-*` Skillを使用）。
- Planning Skillやworkflow policy自体の不具合を直す（`skill-maintainer`を使用）。

## Required Inputs

- `documents/AGENTS.md`
- `documents/.ai/workflow/artifact-map.yaml`
- `documents/.ai/workflow/decision-policy.md`
- `documents/.ai/workflow/state.yaml`
- 対象Featureの`documents/features/{feature_id}/`配下成果物

## Procedure

1. 作業基準を`documents/`へ切り替え、`documents/AGENTS.md`とworkflow stateを読む。
2. stateと成果物を照合し、未完了の最上流工程を特定する。
3. `documents/.agents/skills/planning-orchestrator/SKILL.md`の手順に従い、各Owner Skillを実行するsubagentを起動する。subagentの報告を待ってから次工程へ進む。
4. design-reviewとimplementation-readiness-reviewは、設計担当と異なるfresh subagentへ委譲し、両方の`pass`と利用者の明示承認を待ってからTask Breakdownへ進む。
5. 実装開始前は、対象Featureのfeature-design担当、design-review担当、implementation-readiness-review担当、実装担当の全てと異なるfresh subagentへ`implementation-preflight-review`を委譲し、`pass`を待つ。
6. `pass`後に、対象責務の`impl-*` Skillを使う実装subagentへTaskを委譲する。
7. 実装後は、対象Featureのfeature-design担当、design-review担当、implementation-readiness-review担当、実装担当、実装前reviewerの全てと異なるfresh subagentへ`implementation-conformance-review`を委譲し、`pass`を待つ。`pass`以外は実装完了にしない。
8. Owner SkillのExit Criteriaを確認してからstateを更新し、`implementation-handoff.yaml`完成後だけImplementation領域へ渡す。

## Guardrails

- `documents/.ai/workflow/artifact-map.yaml`のOwnerではない成果物を直接編集しない。
- 利用者の設計承認、または必須の独立レビュー結果を待たずに下流工程へ進めない。
- L3/L4 DecisionはDecision Ownerに記録・提示させ、独自に決定しない。
- Planningに必要なSkillが欠ける、またはworkflow規約に不整合がある場合は`skill-maintainer`へ委譲する。
- 実装前後のreviewerが同じagent、または実装担当agentになる割当てを禁止する。

## Completion Criteria

- 次工程のOwner Skillが特定され、正本workflowに従って委譲されている。
- または対象Featureが`implementation_ready`となり、handoffとstateが整合している。
