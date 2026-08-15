# topic2html Agent Guide

日本語で簡潔かつ丁寧に回答してください。

## 進行の入口

- 要件整理、Feature分割、設計、Task分割、Planningの再開は `topic2html-orchestrator` を使う。正本のworkflow成果物と詳細なPlanning規約は [`documents/AGENTS.md`](documents/AGENTS.md) に従う。
- 実装は、対象Featureの [`documents/features/{feature_id}/implementation-handoff.yaml`](documents/features/{feature_id}/implementation-handoff.yaml) が `implementation_ready` であり、[`documents/.ai/workflow/state.yaml`](documents/.ai/workflow/state.yaml) の`implementation_ready_features`にも当該Featureが記録され、両者が整合する場合だけ開始する。
- 実装では必要に応じて `impl-go-server-postgres`、`impl-react-admin-ui`、`impl-project-verification` を使う。承認済みでない製品・業務・アーキテクチャ・security判断を実装で決めない。

## 役割境界

- Planning成果物のOwnerとDecision Policyは `documents/.ai/workflow/` を正本とする。Planning成果物を実装都合で直接変更しない。
- 実装Skillの不備、配置漏れ、誤発火は `skill-maintainer` で修正する。
- Go Server、管理UI、生成HTML隔離origin、PostgreSQLの責務境界は承認済みDecisionを守る。秘密値・認可情報をBrowser、生成HTML、ログ、fixtureへ出さない。

## ユーザー指摘の継続的改善

ユーザーから成果物、説明、質問、レビュー、進行、判断方法への指摘を受けたら、個別修正だけで足りるか、Skill・workflow規約の再発可能な不足かを判定する。後者は`skill-maintainer`で最小修正とregression scenarioを追加し、前者はSkillを不必要に一般化しない。

## 検証

- 親READMEと、より近い`AGENTS.md`にあるTask・lint・test・build手順を使う。
- 変更前後で関連成果物とdiffを確認し、実行できなかった検証は理由を報告する。
