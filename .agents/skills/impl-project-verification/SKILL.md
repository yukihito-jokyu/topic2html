---
name: impl-project-verification
description: topic2htmlのGo、React、PostgreSQL、外部境界を横断する自動検証を実装・整備する。承認済みcontractのunit、integration、HTTP、E2E検証で使用し、製品の業務仕様や本番資格情報を決める用途には使用しない。
---

# topic2html プロジェクト検証実装

## Goal

承認済みcontractを、Go unit、実PostgreSQL integration、HTTP、Playwright E2Eで再現可能に検証する。通常CIで実Google、Codex、実運用Secretへ接続せず、固定のtest doubleと架空値だけを使う。

## Use When

- Go / React / PostgreSQL / HTTP / 外部OAuth境界をまたぐ検証を追加・更新する。
- migration、repository、認可、公開・隔離境界、画面操作の回帰検証を整備する。
- 統一Taskから実行するlint、test、buildの検証責務を実装・保守する。

## Do Not Use When

- テスト対象のHTTP・DB・業務・security contractそのものを決める。
- 本番Google/Codex資格情報を通常CIやfixtureへ入れる。
- 実装だけで明らかな局所UI変更に、不要な横断テスト基盤を追加する。

## Required Inputs

- 対象Featureのhandoff、受け入れ基準、`design/test-strategy.md`
- 対象HTTP、operation、DB、画面、外部境界contract
- `DEC-ARCH-003.md`、親`AGENTS.md`、既存Task / test設定

## Project Evidence

1. `DEC-ARCH-003`: `npm ci`、Go Module検証、整形・静的検査、unit・integration・browser検証をTaskで再現し、通常CIはGoogle/Codexのtest doubleを使用する。
2. Feature詳細設計: unit、実PostgreSQL integration、HTTP contract、Google test double E2E、秘密情報非露出を層ごとに検証する。
3. 親README: `task verify`を統一検証入口とする。

## Owned Implementation Scope

- unit / integration / HTTP contract / E2Eの責務分割とテスト実装
- 実PostgreSQLを使うmigration・repository結合試験
- 固定discovery、JWKS、Token応答など外部Server test double
- Taskを通じた再現可能な整形、静的検査、build、testの実行
- 秘密情報を含まないfixtureと失敗経路の回帰試験

## File Placement Rules

- 既存のGo、frontend、E2E、Task構成へ従い、テストを対象責務の近くへ置く。
- ProjectのTaskから実行するscriptはリポジトリの`scripts/`へ置き、Skill配下には置かない。
- test doubleとfixtureは本番接続先・実資格情報を必要としない形で管理する。
- E2EがGoogle/Codexや生成HTML隔離originを扱う場合、対象Featureが所有する承認済み境界を超えない。
- Go testは、複数の入力・期待結果・失敗経路を持つ対象をtable-driven testで記述する。
- scripts/check_table_driven_tests.pyでテーブル駆動化の未対応候補を確認する。これは静的な候補検出であり、単一シナリオの並行・資源管理テストは個別に判断する。
- Backendを対象にする場合、実装者と独立reviewerが同じ`task backend:coverage`またはTask規約で明記された同等スクリプトを実行できるようにする。スクリプトは対象範囲の100%未達、計測失敗、test失敗で非0終了にする。

## Autonomous Decisions

- 承認済みcontractの各失敗経路を検証層へ配分する。
- 架空のClient ID、鍵、メール、token、固定応答を使うtest doubleの設計。
- 既存のtest runner、fixture、Taskの慣例に沿う局所的なtest helper。

## Escalate When

- contractに期待結果、失敗時の状態、責任Featureが定義されていない。
- 本番外部サービス・本番Secret・実利用者データへの接続が必要になる。
- 新しいtest framework、CI provider、外部サービスを採用する。
- テスト結果が承認済み設計と矛盾する。

## Procedure

1. handoffとtest strategyから、受け入れ基準を検証層へ対応付ける。
2. 既存test / Task規約を確認し、最小のtest double・fixture・helperを追加する。
3. 正常系だけでなく、期限切れ、再使用、DB障害、外部失敗、未認可、cross-originなどcontract上の失敗経路を検証する。
4. fixture、URL、HTML、ログに実Secret・認可token・許可メールがないことを確認する。
5. 統一Taskと各toolchainの強い検証を実行する。
6. Backendを対象にする場合は、共通カバレッジ検証スクリプトを実行し、100%であることを確認する。
7. 不足contractは実装で補わず、設計差し戻しとして報告する。

## Validation

- `task verify`
- `task backend:coverage`（またはTask規約で明記された同等スクリプト）
- `npm ci`、TypeScript型検査、Biome、production build、Vitest
- `go mod download`、`go mod verify`、`gofmt`、`golangci-lint`、`go test ./...`、Go build
- 実PostgreSQL integrationとPlaywright E2E

## Completion Criteria

- 受け入れ基準が適切な検証層に対応付けられ、失敗経路を含め自動化されている。
- 通常CIで実Google/Codex・実運用Secret・実利用者データを使わない。
- 統一検証から再現可能であり、実行結果を報告できる。
- contract不足を新しいテスト仕様として黙って固定していない。
