# Skill Test Change Log

## 2026-08-14

- 対象: workflow decision policy
- 症状: 日本語で運用するプロジェクトで、Planning成果物が英語で作成された。
- 根本原因: 成果物本文の言語規約がworkflow policyに存在しなかった。
- 修正: 既定のPlanning成果物言語を日本語とする最小規約を追加した。
- Regression Scenario: `workflow-policy/artifact-language-japanese.yaml`

- 対象: planning-orchestrator
- 症状: workflowのBootstrap後、task-breakdownが必須入力として参照する`implementation-handoff-schema.yaml`が存在しなかった。
- 根本原因: Bootstrap対象にhandoff schemaが含まれていなかった。
- 修正: Bootstrap対象へschemaを追加し、既定schemaを定義した。
- Regression Scenario: `planning-orchestrator/bootstrap-handoff-schema.yaml`

- 対象: feature-design
- 症状: Feature固有のL3 Decisionが、命名規約に従わない`design/`配下の補助資料として作成された。
- 根本原因: Feature-local DecisionのID規則はあったが、L3/L4 Decisionの配置・記録規則が明示されていなかった。
- 修正: L3/L4のFeature固有Decisionを`decisions/DEC-FEAT-###.md`に記録し、操作別詳細設計とは分離する規則を追加した。
- Regression Scenario: `feature-design/feature-local-decision-naming.yaml`

## 2026-08-15

- 対象: topic2html-orchestrator / planning-orchestrator
- 症状: オーケストレーターがOwner Skillへの委譲を記していたが、各工程でsubagentを起動して結果を待つ規約が明文化されていなかった。
- 根本原因: ルーティングとsubagent実行の責務が手順として接続されていなかった。
- 修正: Owner Skillごとのsubagent委譲・完了待ちと、独立reviewerのfreshnessを必須化した。
- Regression Scenario: `planning-orchestrator/subagent-owner-routing.yaml`

- 対象: implementation-preflight-review / implementation-conformance-review
- 症状: 実装前のissue・Task・仕様整合確認と、実装後のissue・要件・設計・diff・検証結果の独立監査を行う専用Skillがなかった。
- 根本原因: Planningの実装準備性レビューと、実装後の仕様適合確認の責務が分離されていなかった。
- 修正: 実装開始前と実装後にそれぞれ独立subagentで実行するreview Skillを追加した。
- Regression Scenario: `implementation-preflight-review/preflight-alignment-gate.yaml`、`implementation-conformance-review/conformance-alignment-gate.yaml`

- 対象: implementation-skill-builder
- 症状: Planning成果物が`documents/`、実装Skillが親リポジトリという構成で、親Rootの`AGENTS.md`とPlanning workflowへの入口がないまま実装Skillだけを生成した。
- 根本原因: 実装Skillの配置先を確認しても、実装RootからPlanning正本へ到達する運用入口の確認手順がなかった。
- 修正: 実装Rootが異なる場合に`AGENTS.md`と`planning-orchestrator`を確認し、欠ける場合は`skill-maintainer`で最小の入口を補う規約を追加した。
- Regression Scenario: `implementation-skill-builder/implementation-root-entrypoint.yaml`

- 対象: feature-design / design-review / implementation-readiness-review / task-breakdown
- 症状: 認証APIの設計だけでは利用者が操作する画面の仕様が確定していないのに、画面設計仕様書を作成・監査する規約がなかった。
- 根本原因: Feature Designの分析選択と完了条件が、画面を持つFeatureの画面状態・操作・API対応を独立した成果物として要求していなかった。
- 修正: 画面を新設・変更する、または画面操作を含むFeatureに画面設計仕様書を要求するScreen Design Coverage Gateを追加し、独立レビューとTask分割の監査対象へ接続した。画面を持たない場合は不要理由を記録する。
- Regression Scenario: `feature-design/screen-design-coverage-gate.yaml`、`implementation-readiness-review/screen-design-coverage-gate.yaml`

- 対象: implementation-readiness-review / planning-orchestrator / artifact-map
- 症状: 実装準備性レビューが`human_decision_required`を返しても、利用者に提示するFeature固有Decisionが`decisions/`へ作成されなかった。
- 根本原因: reviewerのOwned OutputとDecision Ownerへのルーティング、およびDecision成果物のartifact ownershipが接続されていなかった。
- 修正: reviewerは必要な判断を記録し、orchestratorはfeature-designへDecision作成をルーティングしてから利用者へ提示する規則と、`features/*/decisions/**`のOwnerを追加した。
- Regression Scenario: `implementation-readiness-review/contract-completeness-gate.yaml`

- 対象: planning-orchestrator / design-review / task-breakdown
- 症状: design-reviewがpassした後、Task Breakdownの監査で初めてDDL、wire contract、設定運用、operation別transactionの不足を検出した。
- 根本原因: 要件整合レビューと、独立した実装者視点の契約完全性監査が同一ゲートではなく、後段のTask Breakdownまで分離されていなかった。
- 修正: design-review後・利用者レビュー前にimplementation-readiness-reviewを必須ゲートとして追加し、Task Breakdownにはそのpassを必須入力として追加した。
- Regression Scenario: `planning-orchestrator/implementation-readiness-gate.yaml`、`design-review/implementation-readiness-next-gate.yaml`、`implementation-readiness-review/contract-completeness-gate.yaml`、`task-breakdown/implementation-readiness-prerequisite.yaml`
