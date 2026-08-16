# Skill Test Change Log

## 2026-08-16

- 対象: task-breakdown
- 症状: 「1 APIを1 Task」の規約がmigrationとCodex app-server adapterまで各API Taskへ含めるよう読め、複数APIから再利用できる横断基盤を独立Taskにできなかった。
- 根本原因: API固有実装の集約と、複数operationに共通する独立検証可能な基盤の分離を区別していなかった。
- 修正: 各APIの固有責務は1 Taskに完結させたまま、複数operationから再利用され、API固有の振る舞いを含まず、単独で検証可能なmigration・Codex app-server adapterなどを横断基盤Taskとして許可する最小規約へ改めた。
- Regression Scenario: `task-breakdown/http-operation-single-task.yaml`

- 対象: task-breakdown
- 症状: HTTP APIの実装を、永続化、外部adapter、usecase、handler、検証という技術層ごとの複数Taskに分け、1 APIを1 Taskで完結させられなかった。
- 根本原因: Task粒度規約が一般的な論理変更単位だけを定め、HTTP API operationを実装単位として集約する規則を持たなかった。
- 修正: HTTP API operationごとに、必要な層とoperation固有の検証を含む1 Taskへ完結させ、技術層による分割を禁止する最小規約を追加した。画面・横断E2EなどAPI operation自体を実装しない責務は別Taskを許可する。
- Regression Scenario: `task-breakdown/http-operation-single-task.yaml`

- 対象: feature-design
- 症状: operation別資料に図があっても、利用者が指定したHTTP API契約資料にAPI全体のフローとClean Architectureの責務境界を確認できなかった。
- 根本原因: Diagram Policyはoperation別図を扱ったが、利用者指定のHTTP契約資料へのAPI全体図の配置と内容を規定していなかった。
- 修正: `design/http-contract.md`があるFeatureで利用者が明示要求した場合、APIフローチャートとClean Architecture視点のシーケンス図を同資料へ記載する最小規約を追加した。
- Regression Scenario: `feature-design/http-contract-api-diagrams.yaml`

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

- 対象: impl-go-server-postgres / impl-project-verification / implementation-conformance-review
- 症状: Backendのカバレッジ100%を、実装者と独立reviewerが同一の機械実行可能な入口で検証する規約がなく、table-driven testの要求も一貫していなかった。
- 根本原因: 実装Skill、検証Skill、独立監査Skillの間で、共通カバレッジスクリプトの作成・実行・非0失敗条件が接続されていなかった。
- 修正: `task backend:coverage`（またはTask規約で明記された同等スクリプト）を共通入口とし、100%未達・計測失敗・test失敗を非0にすること、table-driven test、実装者とfresh reviewer双方の実行を明記した。
- Regression Scenario: `impl-go-server-postgres/backend-coverage-script.yaml`、`impl-project-verification/backend-coverage-script.yaml`、`implementation-conformance-review/backend-coverage-script-gate.yaml`

- 対象: impl-go-server-postgres
- 症状: 承認済み設計が`handler`、`usecase`、`repository`、`domain`と`cmd` composition rootの実ディレクトリを要求していたにもかかわらず、Skillの配置規約は抽象的なClean Architecture表現に留まり、物理配置の移行漏れが再発した。
- 根本原因: `File Placement Rules`にnamed layerごとの配置、許可・禁止依存、設定・Secret読取りの唯一の責務、構造検証が明示されていなかった。
- 修正: named layerの物理配置、依存方向、`cmd`だけでの設定・Secret読取りとconstructor injection、構造検証をSkillへ追加した。
- Regression Scenario: `impl-go-server-postgres/named-layer-physical-layout.yaml`

- 対象: impl-go-server-postgres
- 症状: 承認済みDEC-ARCH-003がGin v1.12.0のHTTP adapter限定、`backend/`/`frontend/`分離、Clean Architecture、`backend/`独立Go moduleへ改定された後も、Skillが標準`net/http`と旧配置を前提にしていた。
- 根本原因: 生成済み実装Skillを承認済みの技術基盤Decisionへ同期する手順が欠け、Decision改定時のSkill driftを検出する回帰シナリオもなかった。
- 修正: Goal、Project Evidence、File Placement RulesをDEC-ARCH-003の境界・依存方向へ最小更新した。
- Regression Scenario: `impl-go-server-postgres/gin-clean-architecture-boundary.yaml`

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
