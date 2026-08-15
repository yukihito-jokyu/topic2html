---
name: impl-go-server-postgres
description: topic2htmlのGo信頼済みServerとPostgreSQL永続化を実装する。HTTP境界、Server限定設定、parameterized SQL、migration、外部Server連携を伴う複数Featureの実装で使用し、React画面だけの変更や製品ルールの再決定には使用しない。
---

# topic2html Go Server・PostgreSQL実装

## Goal

`backend/`の独立Go moduleで動く、GinをHTTP adapterに限定した信頼済みServerとPostgreSQLの責務を、承認済みのClean Architecture設計どおりに実装する。Browser、生成HTML、ログ、fixtureへServer限定の認可情報・秘密値・DB資格情報を露出させない。

## Use When

- Go ServerのHTTP operation、認可境界、外部Server連携、Server限定設定を実装・変更する。
- PostgreSQLの版管理SQL migration、`pgx/v5` / `pgxpool` によるrepository、明示transactionを実装・変更する。
- 複数Featureで再利用される保護記録、公開対象の解決、生成・版・掲載場所のServer側責務を実装する。

## Do Not Use When

- React管理画面だけの状態・表示・アクセシビリティを変更する（`impl-react-admin-ui`を使用）。
- テスト基盤・E2Eだけを構築する（`impl-project-verification`を使用）。
- 要件、業務ルール、公開contract、認証・隔離方式を新たに決める。

## Required Inputs

- 対象Featureの`implementation-handoff.yaml`、`tasks.md`、承認済み詳細設計
- `design/decisions/DEC-ARCH-001.md`、`DEC-ARCH-002.md`、`DEC-ARCH-003.md`
- 親リポジトリの`AGENTS.md`、既存のGo・SQL・Task規約
- 外部境界を扱う場合は対象operation、設定、HTTP、DB、テスト契約

## Project Evidence

1. `AGENTS.md`: 実装用Skillは親リポジトリに置き、承認済みでないfile path・symbol・framework APIをPlanningへ持ち込まない。
2. `DEC-ARCH-003`: `backend/`と`frontend/`を分離し、Backendを独立Go moduleとする。Go 1.26.5、Gin v1.12.0をHTTP adapterに限定したClean Architecture、PostgreSQL、`pgx/v5` / `pgxpool`、parameterized SQL、明示transaction、版管理SQL migrationを採用する。Backendは`handler`、`usecase`、`repository`、`domain`と`cmd` composition rootへ物理的に分ける。Domainはフレームワーク・HTTP・DB・環境変数を参照せず、Usecaseは外部I/Oをportで表現し、Repositoryが外部I/Oを実装する。
3. `DEC-ARCH-001` / `DEC-ARCH-002`: 信頼済みアプリ、生成HTML隔離origin、PostgreSQLの責務を分離する。
4. `backend/`では実行入口が依存を組み立て、Ginはrouting、middleware、request/response変換だけを所有する。FrontendはBackend内部、PostgreSQL、秘密情報へ依存せず、同一originの公開HTTP contractだけに依存する。以後は既存の近接実装を優先する。

## Owned Implementation Scope

- HTTP requestの構文検証、認可・失敗境界、公開可能な応答の組み立て
- Server限定の設定読取りと起動時検証
- `pgxpool`の接続管理、parameterized SQL、明示transaction、repository
- 前方適用・再適用・rollbackを考慮した版管理SQL migration
- Google、Codexなどの外部Server境界。timeout、失敗分類、秘密情報の非露出

## File Placement Rules

- Go実装、Go module、migration、実行設定は`backend/`配下に置き、FrontendのNode.js/TypeScript資産は`frontend/`に分離する。
- `backend/domain/`には業務ルールとdomain型だけを置く。`handler`、`usecase`、`repository`、Gin、HTTP、PostgreSQL、外部SDK、環境変数への依存を持ち込まない。
- `backend/usecase/`にはapplication operationと外部I/O portを置く。`domain`だけへ依存し、`handler`、`repository`、Gin、HTTP、PostgreSQL、外部SDK、環境変数を直接参照しない。
- `backend/repository/`にはPostgreSQL、Google、Codex app-server、設定などの外部I/O adapterを置く。`domain`と`usecase`のportを実装してよいが、`handler`へ依存しない。
- `backend/handler/`にはGin routing、middleware、HTTP request/response変換だけを置く。`usecase`と`domain`へ依存してよいが、`repository`を直接参照しない。
- `backend/cmd/`だけをcomposition rootとし、設定・Secretを読取り、`handler`、`usecase`、`repository`、`domain`をconstructor injectionで組み立てる。設定・Secretをそれ以外の層へ読ませない。
- migrationは版管理SQLとして保存し、アプリ起動やBrowserから任意に適用できる形にしない。
- Secret、実運用token、実メールアドレス、接続文字列をソース、fixture、ログへ置かない。
- Go testは、複数の入力・期待結果・失敗経路を持つ対象をtable-driven testで記述する。
- `backend/`のカバレッジ検証は、実装者と独立reviewerが同じリポジトリ内の入口から実行できるようにする。原則として`task backend:coverage`を提供し、代替入口を使う場合はTask規約に明記する。

## Autonomous Decisions

- 承認済みcontractを満たす局所的なGo package、型、関数、SQLの分割
- repository内の既存命名・error handling・transaction境界への追従
- 明確に可逆なテストdouble、fixture、テスト用設定の選択

## Escalate When

- 新しい業務ルール、公開URI・HTTP表現、永続化schema、migration方針、外部連携方式を必要とする。
- 新しいORM、HTTP framework、DB、認証libraryなど、承認済みでない依存を採用する。
- 信頼済みoriginと生成HTML隔離originの境界、Cookie・CSRF・認可姿勢を変更する。
- 詳細設計と実装上の事実が矛盾する。

## Procedure

1. handoffから対象Taskと依存Taskを確認し、詳細設計のcontractを読み込む。
2. 近接する既存Go・SQL・Task実装を確認する。初期骨格しかない場合は、承認済み技術制約を超える構成を推測しない。
3. Server境界、永続化、migrationを小さく一貫した単位で実装する。
4. 外部入力を検証し、失敗時も秘密情報・保護記録・認可状態を露出または部分更新しないことを確認する。
5. 実装後にimportまたは構造検証を追加・更新し、`handler → usecase → domain`、`repository → usecase/domain`、`cmd → 全層`以外の依存を検出する。
6. unitと実PostgreSQLを使う統合テストを追加・更新する。
7. `task backend:coverage`またはTask規約で定めた同等スクリプトを作成・更新する。backendの対象範囲を計測し、100%未満または計測・テスト失敗時には非0で失敗させる。
8. Validationを実行し、設計不整合はPlanning成果物を直接変更せず報告する。

## Validation

- `task verify`（親リポジトリの統一検証。利用可能な場合）
- `task backend:coverage`（またはTask規約で明記された同等スクリプト）。100%を満たさない場合は完了にしない。
- `go mod download`、`go mod verify`
- `gofmt`検査、`golangci-lint`、`go test ./...`、Go build
- 実PostgreSQLによるmigration / repository結合試験
- `backend/`配下のimportまたは構造検証により、named layerの禁止依存と`cmd`以外の設定・Secret読取りがないこと

## Completion Criteria

- 対象Taskの受け入れ基準と承認済みHTTP・DB・operation contractを満たす。
- SQLはparameterizedで、更新は明示transactionにより整合する。
- 外部失敗・DB障害・認可失敗でfail closedとなり、秘密情報を露出しない。
- 関連する検証結果と、実行できなかった検証の理由を報告する。
