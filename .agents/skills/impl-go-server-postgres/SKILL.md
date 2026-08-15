---
name: impl-go-server-postgres
description: topic2htmlのGo信頼済みServerとPostgreSQL永続化を実装する。HTTP境界、Server限定設定、parameterized SQL、migration、外部Server連携を伴う複数Featureの実装で使用し、React画面だけの変更や製品ルールの再決定には使用しない。
---

# topic2html Go Server・PostgreSQL実装

## Goal

信頼済みアプリoriginで動くGo `net/http` ServerとPostgreSQLの責務を、承認済み設計どおりに実装する。Browser、生成HTML、ログ、fixtureへServer限定の認可情報・秘密値・DB資格情報を露出させない。

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
2. `DEC-ARCH-003`: Go 1.26.5、標準`net/http`、PostgreSQL、`pgx/v5` / `pgxpool`、parameterized SQL、明示transaction、版管理SQL migrationを採用する。
3. `DEC-ARCH-001` / `DEC-ARCH-002`: 信頼済みアプリ、生成HTML隔離origin、PostgreSQLの責務を分離する。
4. 親リポジトリにはまだ実装パターンがない。最初の実装で局所的な構造を選ぶ場合も、これを恒久的な全体規約として断定せず、以後は既存の近接実装を優先する。

## Owned Implementation Scope

- HTTP requestの構文検証、認可・失敗境界、公開可能な応答の組み立て
- Server限定の設定読取りと起動時検証
- `pgxpool`の接続管理、parameterized SQL、明示transaction、repository
- 前方適用・再適用・rollbackを考慮した版管理SQL migration
- Google、Codexなどの外部Server境界。timeout、失敗分類、秘密情報の非露出

## File Placement Rules

- 親リポジトリの既存Go module、migration、Task構成があれば、それに従う。
- 新規骨格では、Go module・SQL migration・実行設定を責務ごとに分けるが、設計が未承認のアプリ全体レイアウトやlibraryをこのSkillだけで採用しない。
- migrationは版管理SQLとして保存し、アプリ起動やBrowserから任意に適用できる形にしない。
- Secret、実運用token、実メールアドレス、接続文字列をソース、fixture、ログへ置かない。

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
5. unitと実PostgreSQLを使う統合テストを追加・更新する。
6. Validationを実行し、設計不整合はPlanning成果物を直接変更せず報告する。

## Validation

- `task verify`（親リポジトリの統一検証。利用可能な場合）
- `go mod download`、`go mod verify`
- `gofmt`検査、`golangci-lint`、`go test ./...`、Go build
- 実PostgreSQLによるmigration / repository結合試験

## Completion Criteria

- 対象Taskの受け入れ基準と承認済みHTTP・DB・operation contractを満たす。
- SQLはparameterizedで、更新は明示transactionにより整合する。
- 外部失敗・DB障害・認可失敗でfail closedとなり、秘密情報を露出しない。
- 関連する検証結果と、実行できなかった検証の理由を報告する。
